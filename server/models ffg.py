from datetime import date
from typing import Literal
import re
from pydantic import ConfigDict, StrictInt, field_validator
from sqlalchemy import CheckConstraint
from sqlmodel import Field, SQLModel

Category = Literal['Городской', 'Горный', 'Шоссейный', 'Детский']
PHOTOS = {'Городской': 'city.png', 'Горный': 'trail.png', 'Шоссейный': 'road.png', 'Детский': 'junior.png'}


class Product(SQLModel, table=True):
    __table_args__ = (
        CheckConstraint('price >= 0 AND rent_per_day >= 0'),
        CheckConstraint("status IN ('available', 'rented', 'sold')"),
    )
    id: int | None = Field(default=None, primary_key=True)
    name: str
    price: int
    rent_per_day: int = 500
    category: str = 'Городской'
    frame_size: str = 'M'
    image: str | None = 'city.png'
    description: str | None = ''
    status: str = 'available'


class ProductInput(SQLModel):
    model_config = ConfigDict(str_strip_whitespace=True)
    name: str = Field(min_length=2, max_length=100)
    price: int = Field(ge=1, le=100_000_000)
    rent_per_day: int = Field(ge=1, le=1_000_000)
    category: Category = 'Городской'
    frame_size: str = Field(default='M', min_length=1, max_length=20)
    description: str = Field(default='', max_length=1000)


class Customer(SQLModel, table=True):
    id: int | None = Field(default=None, primary_key=True)
    name: str
    phone: str = Field(index=True)


class Order(SQLModel, table=True):
    __table_args__ = (
        CheckConstraint("kind IN ('sale', 'rent')"),
        CheckConstraint('amount >= 0'),
    )
    id: int | None = Field(default=None, primary_key=True)
    product_id: int = Field(foreign_key='product.id', index=True)
    customer_id: int = Field(foreign_key='customer.id', index=True)
    kind: str
    start_date: date
    due_date: date | None = None
    returned_date: date | None = None
    amount: int


class ShopMeta(SQLModel, table=True):
    key: str = Field(primary_key=True)
    value: str


class Checkout(SQLModel):
    model_config = ConfigDict(str_strip_whitespace=True)
    name: str = Field(min_length=2, max_length=100)
    phone: str = Field(min_length=7, max_length=32)
    days: StrictInt = Field(default=1, ge=1, le=365)

    @field_validator('phone')
    @classmethod
    def normalize_phone(cls, value):
        if re.search(r'[^0-9+()\s-]', value):
            raise ValueError('Телефон должен содержать только цифры, пробелы, +, - и скобки')
        digits = re.sub(r'\D', '', value)
        if not 7 <= len(digits) <= 15:
            raise ValueError('В телефоне должно быть от 7 до 15 цифр')
        if len(digits) == 11 and digits.startswith('8'):
            digits = '7' + digits[1:]
        return '+' + digits


class OrderView(SQLModel):
    id: int
    product_id: int
    customer_id: int
    product_name: str
    customer_name: str
    phone: str
    kind: str
    start_date: date
    due_date: date | None
    returned_date: date | None
    amount: int
    overdue: bool


class CustomerView(SQLModel):
    id: int
    name: str
    phone: str
    orders_count: int
    total_amount: int
