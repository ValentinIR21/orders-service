package main

import (
	"context"
	"log"
	"net/http"
	"orders/internal/broker"
	"orders/internal/cache"
	"orders/internal/db"
	"orders/internal/handler"
	"os"
)

func main() {

	ctx := context.Background()

	//Инициализия Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient, err := cache.NewConnectionRedis(ctx, redisAddr)
	if err != nil {
		log.Printf("Redis недоступен: %v", err)
		redisClient = nil
	}
	defer redisClient.CloseConnectionRedis()

	//Инициализация PostgreSQL
	postgresURL := os.Getenv("DB_URL")
	if postgresURL == "" {
		postgresURL = "postgres://postgres:pass@localhost:5432/postgres"
	}
	database, err := db.NewConnection(ctx, postgresURL)
	if err != nil {
		log.Fatalf("БД недоступна: %v", err)
	}
	defer database.CloseConnection(ctx)

	// Создание таблицы orders
	if err := database.CreateTable(ctx); err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}
	log.Printf("Таблицы orders создана")

	// Инициализация Kafka
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	if kafkaAddr == "" {
		kafkaAddr = "localhost:9092"
	}
	kafkaBroker, err := broker.NewKafkaBroker([]string{kafkaAddr}, database)
	if err != nil {
		log.Fatal(err)
	}
	defer kafkaBroker.CloseKafka()

	// Запуск консюмера
	go kafkaBroker.GetConsumer("new_orders")

	handler.SetupRouts(database, redisClient, kafkaBroker)

	// Запуск сервера
	if err := http.ListenAndServe(":9091", nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
