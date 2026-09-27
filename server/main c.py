from contextlib import asynccontextmanager
from pathlib import Path
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from sqlmodel import Session
from .database import initialize_database, make_engine
from .models import Checkout, CustomerView, OrderView, Product, ProductInput
from . import services


def create_app(engine=None):
    engine = engine if engine is not None else make_engine()

    @asynccontextmanager
    async def lifespan(app):
        initialize_database(engine)
        yield

    api = FastAPI(title='ВелоТочка — магазин и прокат', version='3.0.0', lifespan=lifespan)
    api.state.engine = engine
    api.mount('/assets', StaticFiles(directory=Path(__file__).resolve().parents[1] / 'assets/photos'), name='assets')

    @api.get('/health', tags=['Состояние'])
    def health():
        return {'status': 'ok', 'name': 'ВелоТочка', 'version': '3.0.0'}

    @api.get('/products', response_model=list[Product], tags=['Каталог'])
    def products():
        return services.catalog(engine)

    @api.get('/products/{product_id}', response_model=Product, tags=['Каталог'])
    def product(product_id: int):
        with Session(engine) as session:
            return services.require_bike(session, product_id)

    @api.post('/products', response_model=Product, status_code=201, tags=['Каталог'])
    def add_product(data: ProductInput):
        return services.save_bike(engine, data)

    @api.put('/products/{product_id}', response_model=Product, tags=['Каталог'])
    def edit_product(product_id: int, data: ProductInput):
        return services.save_bike(engine, data, product_id)

    @api.delete('/products/{product_id}', status_code=204, tags=['Каталог'])
    def remove_product(product_id: int):
        services.delete_bike(engine, product_id)

    @api.post('/products/{product_id}/sale', response_model=OrderView, status_code=201, tags=['Заказы'])
    def sell(product_id: int, data: Checkout):
        return services.place_order(engine, product_id, 'sale', data)

    @api.post('/products/{product_id}/rent', response_model=OrderView, status_code=201, tags=['Заказы'])
    def rent(product_id: int, data: Checkout):
        return services.place_order(engine, product_id, 'rent', data)

    @api.get('/orders', response_model=list[OrderView], tags=['Заказы'])
    def orders():
        return services.list_orders(engine)

    @api.post('/orders/{order_id}/return', response_model=OrderView, tags=['Заказы'])
    def return_bike(order_id: int):
        return services.complete_rental(engine, order_id)

    @api.get('/customers', response_model=list[CustomerView], tags=['Клиенты'])
    def customers():
        return services.list_customers(engine)

    return api


app = create_app()
