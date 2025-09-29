package db

import (
	"fmt"
	"log"
	"time"

	"order-service/internal/config"
	"order-service/internal/order/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)

	gormLogger := logger.Default
	if cfg.DBName != "production" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	//connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Testar conexão
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Conectado ao PostgreSQL")
	return db, nil
}

func Migrate(db *gorm.DB) error {
	log.Println("Executando migrations...")

	err := db.AutoMigrate(
		&model.Order{},
		&model.OrderItem{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Migrations executadas com sucesso")
	return nil
}
