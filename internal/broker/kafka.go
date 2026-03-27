package broker

import (
	"context"
	"encoding/json"
	"log"
	"orders/internal/db"
	"orders/internal/models"

	"github.com/IBM/sarama"
)

type KafkaBroker struct {
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	DB       *db.DB
}

// Иницияализация кафки
func NewKafkaBroker(broker []string, database *db.DB) (*KafkaBroker, error) {

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	prod, err := sarama.NewSyncProducer(broker, config)
	if err != nil {
		return nil, err
	}

	cons, err := sarama.NewConsumer(broker, config)
	if err != nil {
		prod.Close()
		return nil, err
	}

	return &KafkaBroker{
		Producer: prod,
		Consumer: cons,
		DB:       database,
	}, nil
}

// Метод для отправки заказа в очередь
func (kaf *KafkaBroker) PushOrder(topic string, order models.OrderRequest) error {

	bytes, err := json.Marshal(order)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(bytes),
	}
	_, _, err = kaf.Producer.SendMessage(msg)

	return err
}

// Запуск консюмера
func (kaf *KafkaBroker) GetConsumer(topic string) {

	pc, err := kaf.Consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Ошибка косюмера: %v", err)
	}

	go func() {
		for msg := range pc.Messages() {
			var req models.OrderRequest

			if err := json.Unmarshal(msg.Value, &req); err != nil {
				log.Printf("Ошибка парсинга: %v", err)
				continue
			}

			_, err := kaf.DB.InsertOrders(context.Background(), req)
			if err != nil {
				log.Printf("Ошибка записи в БД: %v", err)
			}
		}
	}()
}

// Отключание кафки
func (kaf *KafkaBroker) CloseKafka() error {
	if err := kaf.Producer.Close(); err != nil {
		return err
	}
	return kaf.Consumer.Close()
}
