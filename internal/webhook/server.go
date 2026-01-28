package webhook

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// ServerConfig holds webhook server configuration
type ServerConfig struct {
	Host          string
	Port          int
	WebhookSecret string
	MaxWorkers    int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}

// Server represents the Fiber webhook server
type Server struct {
	app      *fiber.App
	config   *ServerConfig
	handlers *WebhookHandlers
	queue    *JobQueue
}

// NewServer creates a new webhook server
func NewServer(config *ServerConfig, processor JobProcessor) (*Server, error) {
	if config == nil {
		config = &ServerConfig{
			Host:         "0.0.0.0",
			Port:         3000,
			MaxWorkers:   4,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		AppName:      "Cadence Webhook Server",
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Create job queue
	maxWorkers := config.MaxWorkers
	if maxWorkers < 1 {
		maxWorkers = 4
	}
	queue := NewJobQueue(maxWorkers, processor)

	// Create handlers
	handlers := NewWebhookHandlers(config.WebhookSecret, queue, nil)

	// Register routes
	handlers.RegisterRoutes(app)

	return &Server{
		app:      app,
		config:   config,
		handlers: handlers,
		queue:    queue,
	}, nil
}

// Start starts the webhook server
func (s *Server) Start() error {
	// Start job queue
	if err := s.queue.Start(); err != nil {
		return fmt.Errorf("failed to start job queue: %w", err)
	}

	// Start Fiber server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	fmt.Printf("Starting webhook server on %s\n", addr)
	return s.app.Listen(addr)
}

// Stop gracefully stops the webhook server
func (s *Server) Stop() error {
	// Stop job queue first
	if err := s.queue.Stop(); err != nil {
		return fmt.Errorf("failed to stop job queue: %w", err)
	}

	// Shutdown Fiber app
	return s.app.Shutdown()
}

// GetApp returns the Fiber app for testing purposes
func (s *Server) GetApp() *fiber.App {
	return s.app
}

// GetQueue returns the job queue for testing purposes
func (s *Server) GetQueue() *JobQueue {
	return s.queue
}
