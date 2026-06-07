package main

import (
	"backend/auth"
	"backend/campaign"
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/routes"
	"backend/logger"
	"backend/payment"
	"backend/transaction"
	"backend/user"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger.Init()
	defer logger.Sync()

	// Database
	db, err := config.ConnectDB()
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	// Repositories
	userRepository := user.NewRepository(db)
	campaignRepository := campaign.NewRepository(db)
	transactionRepository := transaction.NewRepository(db)

	// Services
	userService := user.NewService(userRepository)
	campaignService := campaign.NewService(campaignRepository)
	authService := auth.NewService()
	paymentService := payment.NewService()
	transactionService := transaction.NewService(transactionRepository, campaignRepository, paymentService)

	// Router
	router := gin.Default()

	// Session store
	cookieStore := cookie.NewStore([]byte(config.SessionSecret()))
	router.Use(sessions.Sessions("mysession", cookieStore))

	// CORS
	router.Use(cors.New(config.CORSConfig()))

	// Rate limiting
	rateLimiter := middleware.NewIPRateLimiter(10, 20)
	router.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Templates & static files
	router.HTMLRender = config.LoadTemplates("./web/templates")
	router.Static("/images", "./images")
	router.Static("/css", "./web/assets/css")
	router.Static("/js", "./web/assets/js")

	// API routes
	api := router.Group("/api/v1")
	routes.RegisterAPI(api, userService, authService, campaignService, transactionService)

	// Web routes (admin panel)
	routes.RegisterWeb(router, userService, campaignService, transactionService)

	// Start server
	port := config.Port()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.Info("server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped gracefully")
}