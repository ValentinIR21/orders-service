package models

import "time"

// Структура таблицы товара
type Order struct {
	ID          int       `json:"id"`
	Product     string    `json:"product"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
}

// Структура запроса
type OrderRequest struct {
	Product     string `json:"product"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}

// Структура ответа
type OrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Order   Order
}
