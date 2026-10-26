package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"

	"api-students/app/model"
	"api-students/config"
)

func main() {
	config.LoadEnv()

	outboxID := flag.Int64("outbox", 0, "ID baris outbox")
	copies := flag.Int("copies", 1, "Jumlah salinan task yang dikirim")
	flag.Parse()

	if *outboxID <= 0 {
		log.Fatal("-outbox harus berupa angka positif")
	}

	ctx := context.Background()
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.GetEnv("DB_USER", "postgres"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "5432"),
		config.GetEnv("DB_NAME", "praktikum_backend"),
		config.GetEnv("DB_SSLMODE", "disable"),
	)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}
	defer conn.Close(ctx)

	var event model.OutboxEvent
	err = conn.QueryRow(ctx,
		`SELECT id, event_type, payload, created_at FROM outbox WHERE id = $1`, *outboxID).
		Scan(&event.ID, &event.EventType, &event.Payload, &event.CreatedAt)
	if err != nil {
		log.Fatalf("mengambil outbox %d: %v", *outboxID, err)
	}

	job := model.NotificationJob{
		OutboxID:  event.ID,
		EventType: event.EventType,
		Payload:   event.Payload,
	}
	payload, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("marshal job: %v", err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	for i := 1; i <= *copies; i++ {
		task := asynq.NewTask(model.TaskSendNotification, payload,
			asynq.Queue(model.QueueNotifications),
			asynq.MaxRetry(5),
			asynq.Timeout(30*time.Second),
			asynq.Retention(24*time.Hour),
		)
		_, err := client.EnqueueContext(ctx, task)
		if err != nil {
			log.Fatalf("gagal enqueue salinan ke-%d: %v", i, err)
		}
	}

	fmt.Printf("berhasil mengirim %d salinan task outbox %d ke queue %s\n",
		*copies, *outboxID, model.QueueNotifications)
}
