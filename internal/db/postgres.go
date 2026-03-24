package db

import (
	"context"
	"fmt"
	"orders/internal/models"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	conn *pgx.Conn
}

// Создания нового подключения к БД
// Принимает контекст и строку подключения
// Возрвращает структуру для работы с БД и error
func NewConnection(ctx context.Context, connString string) (*DB, error) {

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Закрывает соединение с БД
func (db *DB) CloseConnection(ctx context.Context) error {
	return db.conn.Close(ctx)
}

// Создание таблицы
func (db *DB) CreateTable(ctx context.Context) error {

	sqlQuery := `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		product VARCHAR(200) NOT NULL,
		description VARCHAR(1000) NOT NULL,
		price INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.conn.Exec(ctx, sqlQuery)
	return err
}

// Вставка товара в БД
// Принимает контекст и структуру с данными заказа
// Возвращает структуру с полным заказом, включая id and time, error
func (db *DB) InsertOrders(ctx context.Context, orderReq models.OrderRequest) (models.Order, error) {

	var order models.Order

	// Добавление новой строки в таблицу и возврат данных (включая id и create_at)
	sqlQuery := `
	INSERT INTO orders (product, description, price)
	VALUES($1, $2, $3)
	RETURNING id, product, description, price, created_at
	`

	// Заполняем таблицу БД используя QueryRow.
	// Scan - сканируем таблицу и заполняем поля структуры по указателям.
	err := db.conn.QueryRow(ctx, sqlQuery, orderReq.Product, orderReq.Description, orderReq.Price).
		Scan(&order.ID, &order.Product, &order.Description, &order.Price, &order.CreatedAt)

	if err != nil {
		return models.Order{}, fmt.Errorf("Ошибка при вставке: %w", err)
	}

	return order, nil // возвращаем копию структуры
}

// Принимает контекст и возвращает список всех товаров.
func (db *DB) GetAllOrders(ctx context.Context) ([]models.Order, error) {

	sqlQuery := `SELECT id, product, description, price, created_at FROM orders ORDER BY id DESC`

	rows, err := db.conn.Query(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("Ошибка запроса: %w", err)
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.Product, &order.Description, &order.Price, &order.CreatedAt)
		if err != nil {
			continue // Если ошибка в одной строке, то просто пропустим ее пока что
		}

		orders = append(orders, order)
	}

	return orders, nil
}
