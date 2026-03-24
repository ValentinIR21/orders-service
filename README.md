<h1>Простой HTTP сервер на Go с PostgreSQL</h1>
<h2>Возможности:</h2>
- Подключение к PostgresSQL

- Обработка JSON, сохранение в БД и возврат результата
  
- Получение всех заказов
  
- Получение заказа по ID
  
- Кэширование Заказов в Redis
  

<h2>API:</h2>
<h3>Пример запроса JSON:</h3>

{
    "product": "Ноутбук",
    "description": "Игровой ноутбук",
    "price": 75000
}

<h3>Ответ:</h3>

{
    "success": true,
    "message": "Заказ создан",
    "order": {
        "id": 1,
        "product": "Ноутбук",
        "description": "Игровой ноутбук",
        "price": 75000,
        "created_at": "2025-03-23T12:00:00Z"
    }
}

<h2>Запуск:</h2>

**Запустить PostgreSQL**:

- ```docker run -d --name postgres -e POSTGRES_PASSWORD=pass -p 5432:5432 postgres```

**Запустить Redis (необязательно)**:

- ```docker run -d --name redis -p 6379:6379 redis```

**Запустить шарманку**
- ```go run main.go```

  
**запросы**:

- ```http://localhost:9091/order``` - Создание товара (принимает тело в формате JSON)

- ```http://localhost:9091/orders``` - Получение всех товаров

- ```http://localhost:9091/order/{id}``` - Получение товара по id

<h2>План дальнейших действий:</h2>
  
- Kafka consumer для асинхронной обработки
  
- Web-интерфейс для поиска заказов
