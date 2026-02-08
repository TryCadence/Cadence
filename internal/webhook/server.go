package webhook

import (
	"fmt"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/logging"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type ServerConfig struct {
	Host          string
	Port          int
	WebhookSecret string
	MaxWorkers    int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}

type Server struct {
	app      *fiber.App
	config   *ServerConfig
	handlers *WebhookHandlers
	queue    *JobQueue
	log      *logging.Logger
	Cache    analysis.AnalysisCache
	Metrics  analysis.AnalysisMetrics
	Plugins  *analysis.PluginManager
}

func NewServer(config *ServerConfig, processor JobProcessor) (*Server, error) {
	if config == nil {
		config = &ServerConfig{
			Host:         "0.0.0.0",
			Port:         8000,
			MaxWorkers:   4,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 0, // Disabled: SSE streaming requires long-lived responses; handlers use context timeouts instead
		}
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		AppName:      "Cadence Webhook Server",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	maxWorkers := config.MaxWorkers
	if maxWorkers < 1 {
		maxWorkers = 4
	}
	queue := NewJobQueue(maxWorkers, processor)

	handlers := NewWebhookHandlers(config.WebhookSecret, queue, nil)

	// Initialise observability and plugin subsystems
	cache := analysis.NewInMemoryCache(analysis.WithMaxSize(256))
	metrics := analysis.NewInMemoryMetrics()
	plugins := analysis.NewPluginManager()

	handlers.WithCache(cache).WithMetrics(metrics).WithPlugins(plugins)

	handlers.RegisterRoutes(app)

	return &Server{
		app:      app,
		config:   config,
		handlers: handlers,
		queue:    queue,
		log:      logging.Default().With("component", "server"),
		Cache:    cache,
		Metrics:  metrics,
		Plugins:  plugins,
	}, nil
}

// Start starts the webhook server
func (s *Server) Start() error {
	if err := s.queue.Start(); err != nil {
		return fmt.Errorf("failed to start job queue: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.log.Info("starting webhook server", "address", addr, "workers", s.config.MaxWorkers)
	return s.app.Listen(addr)
}

func (s *Server) Stop() error {
	if err := s.queue.Stop(); err != nil {
		return fmt.Errorf("failed to stop job queue: %w", err)
	}

	return s.app.Shutdown()
}

func (s *Server) GetApp() *fiber.App {
	return s.app
}

func (s *Server) GetQueue() *JobQueue {
	return s.queue
}
