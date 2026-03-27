package handler

import (
	"net/http"
	"orders/internal/broker"
	"orders/internal/cache"
	"orders/internal/db"
)

func SetupRouts(db *db.DB, redis *cache.RedisClient, kafka *broker.KafkaBroker) {
	http.HandleFunc("/order", CreateNewOrderHandler(kafka))
	http.HandleFunc("/orders", GetOrdersHandler(db))
	http.HandleFunc("/order/", GetOrderByIdHandler(db, redis))
}
