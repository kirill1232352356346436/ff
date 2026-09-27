Проект реализован в учебной цели изучения Go, и не использовался как-либо.

Для запуска desktop введите в cmd:
cd "путь к проекту"
go run .

Для запуска сервера введите в cmd:
cd "путь к проекту\server"
..venv\Scripts\python -m uvicorn main:app --reload --port 8010

desktop реализован без использования ИИ как-либо
Серверная часть реализована через ИИ

Структура проекта:
backend.go # HTTP-клиент к FastAPI: GET_CATALOG → GET /products → []Config.Buycard
GUI.go # весь интерфейс Fyne: окно, вкладки, каталог, пагинация
Config.go # типы: WindowApp (настройки окна), Buycard (товар)
events.go # каналы Request/Result, SendRequest(), CheckResult()
background1.jpg # фон главного окна
Recycle.jpg # картинка товара на карточках
main.py # FastAPI + SQLModel: модель Product, сид 16 товаров, CRUD
