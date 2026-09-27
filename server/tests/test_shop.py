from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
from tempfile import TemporaryDirectory
import sqlite3
import unittest
from fastapi.testclient import TestClient
from server.database import make_engine, initialize_database
from server.main import create_app

CONTACT = {'name': 'Иван Тестовый', 'phone': '+7 (900) 123-45-67', 'days': 3}


class ShopTest(unittest.TestCase):
    def setUp(self):
        self.temp = TemporaryDirectory()
        self.path = Path(self.temp.name) / 'test.db'
        self.engine = make_engine(self.path)
        self.client = TestClient(create_app(self.engine))
        self.client.__enter__()

    def tearDown(self):
        self.client.__exit__(None, None, None)
        self.engine.dispose()
        self.temp.cleanup()

    def test_new_catalog_has_four_different_photos(self):
        products = self.client.get('/products').json()
        self.assertEqual(len(products), 4)
        self.assertEqual(len({p['image'] for p in products}), 4)
        for p in products:
            self.assertEqual(self.client.get('/assets/' + p['image']).status_code, 200)

    def test_rent_return_sell_and_customer_reuse(self):
        rent = self.client.post('/products/1/rent', json=CONTACT)
        self.assertEqual(rent.status_code, 201, rent.text)
        self.assertEqual(rent.json()['amount'], 1500)
        self.assertEqual(rent.json()['product_name'], 'Сити 28')
        self.assertEqual(self.client.post('/products/1/sale', json=CONTACT).status_code, 409)
        order_id = rent.json()['id']
        self.assertEqual(self.client.post(f'/orders/{order_id}/return').status_code, 200)
        self.assertEqual(self.client.post(f'/orders/{order_id}/return').status_code, 409)
        sale = self.client.post('/products/1/sale', json={**CONTACT, 'phone': '89001234567'})
        self.assertEqual(sale.status_code, 201)
        self.assertEqual(sale.json()['amount'], 28900)
        self.assertEqual(self.client.get('/products/1').json()['status'], 'sold')
        self.assertEqual(self.client.post(f"/orders/{sale.json()['id']}/return").status_code, 409)
        customers = self.client.get('/customers').json()
        self.assertEqual(len(customers), 1)
        self.assertEqual(customers[0]['orders_count'], 2)
        self.assertEqual(customers[0]['total_amount'], 30400)
        self.assertEqual(self.client.delete('/products/1').status_code, 409)

    def test_invalid_contact_and_days_create_nothing(self):
        for changes in ({'name': '  '}, {'phone': '123'}, {'phone': 'abc1234567'}, {'days': 0}, {'days': 366}, {'days': 1.5}):
            result = self.client.post('/products/1/rent', json=CONTACT | changes)
            self.assertEqual(result.status_code, 422, result.text)
        self.assertEqual(self.client.get('/orders').json(), [])
        self.assertEqual(self.client.get('/customers').json(), [])
        self.assertEqual(self.client.get('/products/1').json()['status'], 'available')

    def test_product_crud_and_pricing(self):
        data = dict(name='Новый городской', price=20000, rent_per_day=300, category='Городской', frame_size='L', description='Для города')
        self.assertEqual(self.client.post('/products', json=data | {'price': -1}).status_code, 422)
        created = self.client.post('/products', json=data)
        self.assertEqual(created.status_code, 201, created.text)
        bid = created.json()['id']
        changed = self.client.put(f'/products/{bid}', json=data | {'category': 'Горный', 'price': 21000})
        self.assertEqual(changed.json()['image'], 'trail.png')
        self.assertEqual(changed.json()['price'], 21000)
        self.assertEqual(self.client.delete(f'/products/{bid}').status_code, 204)
        self.assertEqual(self.client.get(f'/products/{bid}').status_code, 404)

    def test_concurrent_checkout_has_one_winner(self):
        def rent(_):
            return self.client.post('/products/2/rent', json=CONTACT).status_code
        with ThreadPoolExecutor(max_workers=2) as pool:
            statuses = list(pool.map(rent, range(2)))
        self.assertCountEqual(statuses, [201, 409])
        self.assertEqual(len(self.client.get('/orders').json()), 1)

    def test_restart_preserves_data(self):
        self.client.post('/products/1/sale', json=CONTACT)
        with TestClient(create_app(self.engine)) as restarted:
            self.assertEqual(len(restarted.get('/orders').json()), 1)
            self.assertEqual(restarted.get('/products/1').json()['status'], 'sold')

    def test_empty_catalog_is_not_seeded_again(self):
        for bid in range(1, 5):
            self.assertEqual(self.client.delete(f'/products/{bid}').status_code, 204)
        initialize_database(self.engine)
        self.assertEqual(self.client.get('/products').json(), [])

    def test_old_database_migration(self):
        legacy = Path(self.temp.name) / 'legacy.db'
        with sqlite3.connect(legacy) as con:
            con.execute('CREATE TABLE product (id INTEGER PRIMARY KEY, name TEXT NOT NULL, price INTEGER NOT NULL, image TEXT, description TEXT)')
            con.execute("INSERT INTO product VALUES (1, 'Старый велосипед', 12345, 'BuyCard/Recycle.jpg', 'Описание')")
        old_engine = make_engine(legacy)
        try:
            with TestClient(create_app(old_engine)) as old_client:
                rows = old_client.get('/products').json()
                self.assertEqual(len(rows), 1)
                self.assertEqual(rows[0]['name'], 'Старый велосипед')
                self.assertEqual(rows[0]['price'], 12345)
                self.assertEqual(rows[0]['image'], 'city.png')
                self.assertEqual(rows[0]['status'], 'available')
        finally:
            old_engine.dispose()


if __name__ == '__main__':
    unittest.main()
