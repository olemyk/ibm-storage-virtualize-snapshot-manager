package main

import (
	"fmt"
	"log"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/api"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/auth"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/config"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/db"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification/providers"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/scheduler"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/svc"
)

type appServices struct {
	authService         *auth.Service
	svcClient           *svc.Client
	snapshotScheduler   *scheduler.Scheduler
	notificationManager *notification.Manager
	notifier            *notification.Notifier
	server              *api.Server
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	database, err := initDatabase(cfg)
	if err != nil {
		return err
	}
	defer database.Close()

	services, err := initServices(cfg, database)
	if err != nil {
		return err
	}

	if services.notifier != nil {
		defer services.notifier.Stop()
	}

	if err := services.snapshotScheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	defer services.snapshotScheduler.Stop()

	log.Printf("Starting server on %s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := services.server.Start(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return cfg, nil
}

func initDatabase(cfg *config.Config) (*db.DB, error) {
	database, err := db.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := database.Initialize(); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	log.Println("Database initialized successfully")
	return database, nil
}

func initServices(cfg *config.Config, database *db.DB) (*appServices, error) {
	notificationManager, notifier := initNotificationComponents(database, cfg.Security.EncryptionKey)

	authService := auth.NewService(cfg.Security.JWTSecret)
	svcClient := svc.NewClient()
	snapshotScheduler := scheduler.New(database, svcClient, cfg.Security.EncryptionKey)

	if notifier != nil {
		snapshotScheduler.SetNotifier(notifier)
		log.Println("Notifier connected to scheduler")
	}

	server := api.NewServer(
		cfg,
		database,
		authService,
		svcClient,
		snapshotScheduler,
		notificationManager,
		notifier,
	)

	return &appServices{
		authService:         authService,
		svcClient:           svcClient,
		snapshotScheduler:   snapshotScheduler,
		notificationManager: notificationManager,
		notifier:            notifier,
		server:              server,
	}, nil
}

func initNotificationComponents(database *db.DB, encryptionKey string) (*notification.Manager, *notification.Notifier) {
	providers.RegisterFactory()

	notificationManager := notification.NewManager(database, encryptionKey)
	if err := notificationManager.LoadChannels(); err != nil {
		log.Printf("Warning: Failed to load notification channels: %v", err)
	}

	notifier, err := notification.NewNotifier(database, encryptionKey)
	if err != nil {
		log.Printf("Warning: Failed to create notifier: %v", err)
		return notificationManager, nil
	}

	notifier.Start()
	log.Println("Notification system started")

	return notificationManager, notifier
}

//
