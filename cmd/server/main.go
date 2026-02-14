package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cracked-pm/internal/handler"
	"cracked-pm/internal/store"
	"cracked-pm/migrations"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func main() {
	// Load .env (ignore error — file is optional in production).
	_ = godotenv.Load()

	port := envOr("PORT", "3000")
	databaseURL := envRequired("DATABASE_URL")

	// Connect to PostgreSQL.
	db, err := store.Connect(databaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Run embedded migrations.
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(db.DB, "."); err != nil {
		log.Fatalf("goose up: %v", err)
	}

	// Create store and handler.
	s := store.New(db)
	h := handler.New(s)

	// Create Fiber app.
	app := fiber.New(fiber.Config{
		AppName: "cracked-pm",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	// Static files.
	app.Get("/static/*", static.New("./static"))

	// Routes.
	app.Get("/", h.HandleHome)
	app.Patch("/tasks/:id/move", h.HandleMoveTask)

	// Graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	log.Printf("server started on :%s", port)

	<-ctx.Done()
	log.Println("shutting down...")
	_ = app.Shutdown()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
