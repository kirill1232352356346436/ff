from contextlib import contextmanager
from pathlib import Path
import os
from sqlalchemy import event
from sqlmodel import SQLModel, Session, create_engine, select
from .models import Product, ShopMeta, PHOTOS
from .catalog_seed import SEED


def make_engine(path=None):
    database_path = Path(path or os.getenv('VELO_DB_PATH') or Path(__file__).with_name('store.db'))
    database_path.parent.mkdir(parents=True, exist_ok=True)
    engine = create_engine(f'sqlite:///{database_path}', connect_args={'check_same_thread': False, 'timeout': 10})

    @event.listens_for(engine, 'connect')
    def configure_sqlite(connection, _):
        connection.execute('PRAGMA foreign_keys=ON')
        connection.execute('PRAGMA busy_timeout=10000')
    return engine


def initialize_database(engine):
    SQLModel.metadata.create_all(engine)
    # create_all не изменяет существующие таблицы, поэтому поля добавляются отдельно.
    with engine.begin() as connection:
        columns = {row[1] for row in connection.exec_driver_sql('PRAGMA table_info(product)')}
        additions = {
            'rent_per_day': 'INTEGER NOT NULL DEFAULT 500',
            'status': "TEXT NOT NULL DEFAULT 'available'",
            'category': "TEXT NOT NULL DEFAULT 'Городской'",
            'frame_size': "TEXT NOT NULL DEFAULT 'M'",
        }
        for name, definition in additions.items():
            if name not in columns:
                connection.exec_driver_sql(f'ALTER TABLE product ADD COLUMN {name} {definition}')
    with write_session(engine) as session:
        if session.get(ShopMeta, 'catalog_initialized') is None:
            if session.exec(select(Product)).first() is None:
                session.add_all(Product(**item) for item in SEED)
            session.add(ShopMeta(key='catalog_initialized', value='1'))
        for bike in session.exec(select(Product)):
            if not bike.image or 'BuyCard/' in bike.image or 'Recycle.jpg' in bike.image:
                bike.image = PHOTOS.get(bike.category, 'city.png')


@contextmanager
def write_session(engine):
    # Проверка наличия и изменение статуса выполняются под одной блокировкой.
    with Session(engine, expire_on_commit=False) as session:
        session.connection().exec_driver_sql('BEGIN IMMEDIATE')
        try:
            yield session
            session.commit()
        except Exception:
            session.rollback()
            raise
