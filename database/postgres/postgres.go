package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gormDB *gorm.DB

func InitializePostgres() {
	dns := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DBNAME"))

	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Info),
		NowFunc: time.Now().UTC,
	})

	if err != nil {
		panic("failed to connect database")
	}

	gormDB = db
}

func GetDb(ctx *context.Context) *gorm.DB {
	return gormDB.WithContext(*ctx)
}

func ExecuteQueries(ctx *context.Context, queries ...string) error {
	db := GetDb(ctx)

	db.Transaction(func(tx *gorm.DB) error {
		for _, query := range queries {
			result := tx.Exec(query)

			if result.Error != nil {
				log.Println(result.Error)
				return result.Error
			}
		}
		return nil
	})

	return nil
}
