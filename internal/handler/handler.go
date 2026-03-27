package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"orders/internal/broker"
	"orders/internal/cache"
	"orders/internal/db"
	"orders/internal/models"
	"strconv"
	"time"
)

// Создание нового товара
func CreateNewOrderHandler(kafka *broker.KafkaBroker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Ожидается метод POST", http.StatusMethodNotAllowed)
			return
		}

		var req models.OrderRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Не прошло валидацию", http.StatusBadRequest)
			return
		}

		if req.Product == "" || req.Description == "" || req.Price <= 0 {
			http.Error(w, "Не прошло валидацию", http.StatusBadRequest)
			return
		}

		if err := kafka.PushOrder("new_orders", req); err != nil {
			log.Printf("Ошибка отправки в kafka: %v", err)
			http.Error(w, "Ошибка отправки в kafka", http.StatusInternalServerError)
			return
		}
		log.Printf("Сообщение отправлено в kafka")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "accept",
			"message": "Товар принят",
		})
	}
}

// Получаение товаров
func GetOrdersHandler(db *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Ожидается метод GET", http.StatusMethodNotAllowed)
			return
		}

		orders, err := db.GetAllOrders(r.Context())
		if err != nil {
			http.Error(w, "Ошибка получения товара", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(orders)
	}
}

// Получение по ID товара
func GetOrderByIdHandler(db *db.DB, redis *cache.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Ожидается метод GET", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Path[len("/order/"):]
		if idStr == "" {
			http.Error(w, "ID не указан", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "ID должен быть числом", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		var order models.Order

		if redis != nil {
			val, err := redis.Client.Get(ctx, idStr).Result()
			if err == nil {
				if err := json.Unmarshal([]byte(val), &order); err == nil {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(order)
					log.Printf("Заказ %d из Redis", id)
					return
				}
			}
		}

		order, err = db.GetOrderById(ctx, id)
		if err != nil {
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}

		if redis != nil {
			data, _ := json.Marshal(order)
			redis.Client.Set(ctx, idStr, data, 1*time.Hour)
			log.Printf("Заказ %d добавлен в Redis", id)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}
