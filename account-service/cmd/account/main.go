package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/abhiii71/engineering/account-service/config"
	"github.com/abhiii71/engineering/account-service/internal"
	"github.com/joho/godotenv"
	"github.com/tinrab/retry"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env file not found")
	}

	dbURL := config.DatabaseURL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	var repository internal.AccountRepository

	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		db, err := sql.Open("pgx", dbURL)
		if err != nil {
			log.Println("DB connection error:", err)
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			log.Println("DB ping error:", err)
			return err
		}

		repository = internal.NewAccountRepository(db)
		return nil
	})

	defer repository.Close()

	port := config.Port
	svc := internal.NewService(repository)
	log.Fatal(internal.ListenREST(svc, port))
}
