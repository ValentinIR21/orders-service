<h1>Простой HTTP сервер на Go с PostgreSQL</h1>
<h2>Возможности:</h2>
- Подключение к PostgresSQL
- Обработка JSON, сохранение в БД и возврат результата

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

- ```docker run -d --name postgres -e POSTGRES_PASSWORD=pass -p 5432:5432 postgres```
- ```go run main.go```

  
*запросы*:

- ```http://localhost:9091/order```

- ```http://localhost:9091/orders```

<h2>План дальнейших действий:</h2>

- Добавить handler для получения всех заказов +

- Добавить handler для получения заказа по id

- Добавить кэш (map + RWMutex)
  
- Kafka consumer для асинхронной обработки
  
- Web-интерфейс для поиска заказов
