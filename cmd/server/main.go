package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"device-assignment-api/internal/config"
	"device-assignment-api/internal/database"
	"device-assignment-api/internal/handlers"
	"device-assignment-api/internal/middleware"
	"device-assignment-api/internal/services"
	"device-assignment-api/pkg/auth"
	"device-assignment-api/pkg/logger"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize logger
	log := logger.New()
	log.Info("Starting Device Assignment API Service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	log.Info("Configuration loaded successfully")

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info("Database connection established")

	// Run database migrations
	if err := db.RunMigrations(); err != nil {
		log.Error("Failed to run database migrations", "error", err)
		os.Exit(1)
	}

	log.Info("Database migrations completed")

	// Initialize repositories
	deviceRepo := database.NewDeviceRepository(db.DB())
	assignmentRepo := database.NewAssignmentRepository(db.DB())

	// Initialize services
	deviceService := services.NewDeviceService(deviceRepo, assignmentRepo, log)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.TokenDuration, cfg.JWT.Issuer)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTAuthMiddleware(jwtManager, log)
	certMiddleware := middleware.NewCertificateAuthMiddleware(log)

	// Initialize handlers
	deviceHandler := handlers.NewDeviceHandler(deviceService, log)

	// Setup routes
	router := setupRoutes(deviceHandler, jwtMiddleware, certMiddleware, log)

	// Configure TLS
	tlsConfig, err := configureTLS(&cfg.TLS)
	if err != nil {
		log.Error("Failed to configure TLS", "error", err)
		os.Exit(1)
	}

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		TLSConfig:    tlsConfig,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Starting HTTPS server", "port", cfg.Server.Port)
		if err := server.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Give server 30 seconds to gracefully shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Server exited gracefully")
}

// setupRoutes configures the HTTP routes
func setupRoutes(
	deviceHandler *handlers.DeviceHandler,
	jwtMiddleware *middleware.JWTAuthMiddleware,
	certMiddleware *middleware.CertificateAuthMiddleware,
	log logger.Logger,
) *mux.Router {
	router := mux.NewRouter()

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Device authentication endpoint (requires client certificate)
	api.Handle("/devices/authenticate",
		certMiddleware.Authenticate(http.HandlerFunc(deviceHandler.AuthenticateDevice))).
		Methods("POST")

	// Device management endpoints (require JWT authentication)
	api.Handle("/devices/{deviceId}",
		jwtMiddleware.Authenticate(http.HandlerFunc(deviceHandler.GetDevice))).
		Methods("GET")

	api.Handle("/devices/{deviceId}/assign",
		jwtMiddleware.Authenticate(http.HandlerFunc(deviceHandler.AssignDevice))).
		Methods("POST")

	api.Handle("/devices/{deviceId}/unassign",
		jwtMiddleware.Authenticate(http.HandlerFunc(deviceHandler.UnassignDevice))).
		Methods("DELETE")

	api.Handle("/users/me/devices",
		jwtMiddleware.Authenticate(http.HandlerFunc(deviceHandler.GetUserDevices))).
		Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Log all routes
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			methods, _ := route.GetMethods()
			log.Debug("Registered route", "path", pathTemplate, "methods", methods)
		}
		return nil
	})

	return router
}

// configureTLS sets up TLS configuration for mTLS
func configureTLS(tlsConfig *config.TLSConfig) (*tls.Config, error) {
	// Load server certificate
	cert, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
		PreferServerCipherSuites: true,
	}

	// Load CA certificate for client verification if provided
	if tlsConfig.CAFile != "" {
		caCert, err := os.ReadFile(tlsConfig.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := config.ClientCAs
		if caCertPool == nil {
			caCertPool = x509.NewCertPool()
		}

		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		config.ClientCAs = caCertPool
	}

	return config, nil
}
