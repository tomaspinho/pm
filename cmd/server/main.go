package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cracked-pm/internal/handler"
	"cracked-pm/internal/middleware"
	"cracked-pm/internal/store"
	"cracked-pm/migrations"
	"cracked-pm/views"

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
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

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

	// Create Fiber app with custom error handler.
	app := fiber.New(fiber.Config{
		AppName: "cracked-pm",
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			if code == fiber.StatusForbidden {
				c.Set("Content-Type", "text/html; charset=utf-8")
				return views.ForbiddenPage().Render(c.Context(), c.Response().BodyWriter())
			}

			// Default Fiber error handler behavior.
			c.Set("Content-Type", "text/plain; charset=utf-8")
			return c.Status(code).SendString(err.Error())
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())

	// Static files.
	app.Get("/static/*", static.New("./static"))

	// --- Public routes (no auth) ---
	app.Get("/login", h.HandleShowLogin)
	app.Post("/login", h.HandleLogin)
	app.Get("/signup", h.HandleShowSignup)
	app.Post("/signup", h.HandleSignup)

	// --- Authenticated routes ---
	authed := app.Group("", middleware.RequireAuth(s))

	authed.Post("/logout", h.HandleLogout)
	authed.Get("/", h.HandleLandingPage)
	authed.Get("/projects", h.HandleProjectPicker)

	// --- Org-scoped routes (auth + membership check) ---
	org := authed.Group("/orgs/:org_id", middleware.RequireOrgAccess(s))

	// Project creation.
	org.Get("/projects/new", h.HandleShowCreateProject)
	org.Post("/projects", h.HandleCreateProject)

	// Board view.
	org.Get("/projects/:project_id", h.HandleBoard)

	// Task routes.
	tasks := org.Group("/projects/:project_id/tasks")
	tasks.Get("/new-form", h.HandleNewTaskForm)
	tasks.Get("/cancel-form", h.HandleCancelForm)
	tasks.Post("/", h.HandleCreateTask)
	tasks.Patch("/:id/move", h.HandleMoveTask)
	tasks.Delete("/:id", h.HandleDeleteTask)
	tasks.Get("/:id/detail", h.HandleTaskDetail)
	tasks.Patch("/:id", h.HandleUpdateTask)
	tasks.Patch("/:id/status", h.HandleUpdateStatus)
	tasks.Post("/:id/dependencies", h.HandleAddDependency)
	tasks.Delete("/:id/dependencies/:depID", h.HandleRemoveDependency)
	tasks.Post("/:id/metadata", h.HandleAddMetadata)
	tasks.Patch("/:id/metadata/:oldKey", h.HandleUpdateMetadata)
	tasks.Delete("/:id/metadata/:key", h.HandleDeleteMetadata)

	// Comment routes
	tasks.Post("/:id/comments", h.HandleCreateComment)
	tasks.Post("/:id/comments/:parent_id/reply", h.HandleReplyToComment)
	tasks.Patch("/:id/comments/:comment_id", h.HandleUpdateComment)
	tasks.Delete("/:id/comments/:comment_id", h.HandleDeleteComment)

	// Background session cleanup (every hour).
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	go sessionCleanup(cleanupCtx, s)

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
	cleanupCancel()
	_ = app.Shutdown()
}

// sessionCleanup periodically removes expired sessions from the database.
func sessionCleanup(ctx context.Context, s *store.Store) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := s.DeleteExpiredSessions(ctx)
			if err != nil {
				log.Printf("session cleanup error: %v", err)
			} else if count > 0 {
				log.Printf("session cleanup: removed %d expired sessions", count)
			}
		}
	}
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
