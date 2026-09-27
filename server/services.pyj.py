"""Продажа, аренда, возврат, клиенты и целостность истории заказов."""
from datetime import date, timedelta
from fastapi import HTTPException
from sqlmodel import Session, select
from .database import write_session
from .models import Checkout, Customer, CustomerView, Order, OrderView, PHOTOS, Product, ProductInput


def require_bike(session, product_id):
    bike = session.get(Product, product_id)
    if bike is None:
        raise HTTPException(404, 'Велосипед не найден')
    return bike


def order_view(session, order):
    bike = require_bike(session, order.product_id)
    client = session.get(Customer, order.customer_id)
    return OrderView(**order.model_dump(), product_name=bike.name,
                     customer_name=client.name, phone=client.phone,
                     overdue=bool(order.kind == 'rent' and order.returned_date is None
                                  and order.due_date and order.due_date < date.today()))


def catalog(engine):
    with Session(engine) as session:
        return list(session.exec(select(Product).order_by(Product.id)))


def save_bike(engine, data: ProductInput, product_id=None):
    with write_session(engine) as session:
        bike = require_bike(session, product_id) if product_id is not None else Product(**data.model_dump())
        for key, value in data.model_dump().items():
            setattr(bike, key, value)
        bike.image = PHOTOS[data.category]
        session.add(bike)
        session.flush()
        result = Product.model_validate(bike.model_dump())
    return result


def delete_bike(engine, product_id):
    with write_session(engine) as session:
        bike = require_bike(session, product_id)
        if session.exec(select(Order).where(Order.product_id == product_id)).first():
            raise HTTPException(409, 'Удаление невозможно: у велосипеда есть история заказов')
        session.delete(bike)


def place_order(engine, product_id: int, kind: str, data: Checkout):
    with write_session(engine) as session:
        bike = require_bike(session, product_id)
        if bike.status != 'available':
            raise HTTPException(409, 'Велосипед уже продан или находится в аренде. Обновите каталог.')
        customer = session.exec(select(Customer).where(Customer.phone == data.phone).order_by(Customer.id)).first()
        if customer is None:
            customer = Customer(name=data.name, phone=data.phone)
            session.add(customer)
            session.flush()
        order = Order(product_id=product_id, customer_id=customer.id, kind=kind,
                      start_date=date.today(),
                      due_date=date.today() + timedelta(days=data.days) if kind == 'rent' else None,
                      amount=bike.rent_per_day * data.days if kind == 'rent' else bike.price)
        session.add(order)
        bike.status = 'rented' if kind == 'rent' else 'sold'
        session.flush()
        result = order_view(session, order)
    return result


def complete_rental(engine, order_id):
    with write_session(engine) as session:
        order = session.get(Order, order_id)
        if order is None:
            raise HTTPException(404, 'Заказ не найден')
        if order.kind != 'rent' or order.returned_date is not None:
            raise HTTPException(409, 'Возврат доступен только для действующей аренды')
        order.returned_date = date.today()
        require_bike(session, order.product_id).status = 'available'
        session.flush()
        result = order_view(session, order)
    return result


def list_orders(engine):
    with Session(engine) as session:
        return [order_view(session, row) for row in session.exec(select(Order).order_by(Order.id.desc()))]


def list_customers(engine):
    with Session(engine) as session:
        totals = {}
        for order in session.exec(select(Order)):
            count, amount = totals.get(order.customer_id, (0, 0))
            totals[order.customer_id] = (count + 1, amount + order.amount)
        return [CustomerView(**customer.model_dump(), orders_count=totals.get(customer.id, (0, 0))[0],
                             total_amount=totals.get(customer.id, (0, 0))[1])
                for customer in session.exec(select(Customer).order_by(Customer.id.desc()))]
