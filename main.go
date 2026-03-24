package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orders/internal/cache"
	"orders/internal/db"
	"orders/internal/models"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var database *db.DB
var redisClient *redis.Client

func main() {

	ctx := context.Background()

	connString := "postgres://postgres:pass@localhost:5432/postgres"

	var err error

	redisClient, err = cache.NewConnectionRedis(ctx)
	if err != nil {
		log.Printf("Redis недоступен: %v", err)
		redisClient = nil
	}

	database, err = db.NewConnection(ctx, connString)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer database.CloseConnection(ctx)

	if err := database.CreateTable(ctx); err != nil {
		log.Fatal("Ошибка при создании таблицы:", err)
	}
	fmt.Println("Успешное создание таблицы orders")

	http.HandleFunc("/order", createOrderHandler)
	http.HandleFunc("/orders", getOrders)
	http.HandleFunc("/order/", getOrderId)

	if err := http.ListenAndServe(":9091", nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}

}

func getOrderId(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Ожидается метод: GET", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/order/"):]
	if idStr == "" {
		http.Error(w, "ID заказа не указан", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Ожидается получение id в виде числа", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	var order models.Order

	// Добавление товара в redis
	if redisClient != nil {
		val, err := redisClient.Get(ctx, idStr).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(val), &order); err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(order)
				log.Printf("Товар %q найден в Redis", idStr)
				return
			}
		}
	}

	// Ищем в БД, если нет в Redis
	order, err = database.GetOrderById(ctx, id)
	if err != nil {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	// Добавляем товар в Redis на часик
	if redisClient != nil {
		data, _ := json.Marshal(order)
		if err := redisClient.Set(ctx, idStr, data, 1*time.Hour).Err(); err != nil {
			log.Printf("Ошибка при сохранении в Redis: %v", err)
		}
		log.Printf("Товар %q добавлен в Redis", idStr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)

	log.Printf("запрос '/orders/{id}' отработал успешно")

}

func getOrders(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Ожидается метод: GET", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	var orders []models.Order

	orders, err := database.GetAllOrders(ctx)
	if err != nil {
		http.Error(w, "Ошибка получения товаров из БД:"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)

	log.Printf("запрос '/orders' отработал успешно")
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Ожидается метод: POST", http.StatusMethodNotAllowed)
		return
	}

	var req models.OrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Не верный JSON:"+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Product == "" {
		http.Error(w, "Поле product обязательное для заполнения", http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, "Поле description обязательное для заполнения", http.StatusBadRequest)
		return
	}

	if req.Price <= 0 {
		http.Error(w, "сумма не может быть отрицательное или меньше 0", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	order, err := database.InsertOrders(ctx, req)
	if err != nil {
		http.Error(w, "ошибка сохранения"+err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.OrderResponse{
		Success: true,
		Message: "Заказ создан",
		Order:   order,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	log.Printf("Товар принят")

}
