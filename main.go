package main

import (
	"backend/auth"
	"backend/campaign"
	"backend/handler"
	"backend/helper"
	"backend/logger"
	"backend/middleware"
	"backend/payment"
	"backend/transaction"
	"backend/user"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	webHandler "backend/web/handler"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// func init() {

// 	err := godotenv.Load(".env")
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// }

func main() {
	logger.Init()
	defer logger.Sync()

	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_NAME := os.Getenv("DB_NAME")
	DB_PORT := os.Getenv("DB_PORT")
	DB_URL := os.Getenv("DB_URL")

	dsn := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=require TimeZone=Asia/Shanghai", DB_URL, DB_USER, DB_PASSWORD, DB_PORT, DB_NAME)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	userRepository := user.NewRepository(db)
	campaignRepository := campaign.NewRepository(db)
	transactionRepository := transaction.NewRepository(db)

	userService := user.NewService(userRepository)
	campaignService := campaign.NewService(campaignRepository)
	authService := auth.NewService()
	paymentService := payment.NewService()
	transactionService := transaction.NewService(transactionRepository, campaignRepository, paymentService)

	userHandler := handler.NewUserHandler(userService, authService)
	campaignHandler := handler.NewCampaignHandler(campaignService)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	userWebHandler := webHandler.NewUserHandler(userService)
	campaignWebHandler := webHandler.NewCampaignHandler(campaignService, userService)
	transactionWebHandler := webHandler.NewTransactionHandler(transactionService)
	sessionWebHandler := webHandler.NewSessionHandler(userService)

	router := gin.Default()
	// router.Use(cors.Default())

	sessionSecret := os.Getenv("SESSION_SECRET_KEY")
	if sessionSecret == "" {
		sessionSecret = "CHANGE_ME_SESSION_KEY"
	}
	cookieStore := cookie.NewStore([]byte(sessionSecret))
	router.Use(sessions.Sessions("mysession", cookieStore))
	// CORS configuration — configure ALLOWED_ORIGINS env var for production
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	corsOrigins := []string{"http://localhost:3000", "http://localhost:8080"}
	if allowedOrigins != "" {
		corsOrigins = strings.Split(allowedOrigins, ",")
	}
	corsConfig := cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	// Allow all origins only when no ALLOWED_ORIGINS is set and AllowCredentials is false
	if os.Getenv("ALLOWED_ORIGINS") == "" && os.Getenv("ENV") == "development" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false
	}
	router.Use(cors.New(corsConfig))

	// Rate limiting: 10 requests per second, burst of 20
	rateLimiter := middleware.NewIPRateLimiter(10, 20)
	router.Use(middleware.RateLimitMiddleware(rateLimiter))

	router.HTMLRender = loadTemplates("./web/templates")

	router.Static("/images", "./images")
	router.Static("/css", "./web/assets/css")
	router.Static("/js", "./web/assets/js")

	api := router.Group("/api/v1")

	api.POST("/users", userHandler.RegisterUser)
	api.POST("/sessions", userHandler.LoginUser)
	api.POST("/email_checkers", userHandler.CheckEmailAvailability)
	api.POST("/avatars", authMiddleware(authService, userService), userHandler.UploadAvatar)
	api.GET("/users/fetch", authMiddleware(authService, userService), userHandler.FetchUser)

	api.GET("/campaigns", campaignHandler.GetCampaigns)
	api.GET("/campaigns/:id", campaignHandler.GetCampaign)
	api.POST("/campaigns", authMiddleware(authService, userService), campaignHandler.CreateCampaign)
	api.PUT("/campaigns/:id", authMiddleware(authService, userService), campaignHandler.UpdateCampaign)
	api.POST("/campaign-images", authMiddleware(authService, userService), campaignHandler.UploadImage)

	api.GET("/campaigns/:id/transactions", authMiddleware(authService, userService), transactionHandler.GetCampaignTransactions)
	api.GET("/transactions", authMiddleware(authService, userService), transactionHandler.GetUserTransactions)
	api.POST("/transactions", authMiddleware(authService, userService), transactionHandler.CreateTransaction)
	api.POST("/transactions/notification", transactionHandler.GetNotification)

	router.GET("/users", authAdminMiddleware(), userWebHandler.Index)
	router.GET("/users/new", authAdminMiddleware(), userWebHandler.New)
	router.POST("/users", authAdminMiddleware(), userWebHandler.Create)
	router.GET("/users/edit/:id", authAdminMiddleware(), userWebHandler.Edit)
	router.POST("/users/update/:id", authAdminMiddleware(), userWebHandler.Update)
	router.GET("/users/delete/:id", authAdminMiddleware(), userWebHandler.Delete)
	router.GET("/users/avatar/:id", authAdminMiddleware(), userWebHandler.NewAvatar)
	router.POST("/users/avatar/:id", authAdminMiddleware(), userWebHandler.CreateAvatar)

	router.GET("/campaigns", authAdminMiddleware(), campaignWebHandler.Index)
	router.GET("/campaigns/new", authAdminMiddleware(), campaignWebHandler.New)
	router.POST("/campaigns", authAdminMiddleware(), campaignWebHandler.Create)
	router.GET("/campaigns/image/:id", authAdminMiddleware(), campaignWebHandler.NewImage)
	router.POST("/campaigns/image/:id", authAdminMiddleware(), campaignWebHandler.CreateImage)
	router.GET("/campaigns/edit/:id", authAdminMiddleware(), campaignWebHandler.Edit)
	router.POST("/campaigns/update/:id", authAdminMiddleware(), campaignWebHandler.Update)
	router.GET("/campaigns/show/:id", authAdminMiddleware(), campaignWebHandler.Show)

	router.GET("/transactions", authAdminMiddleware(), transactionWebHandler.Index)

	router.GET("/login", sessionWebHandler.Index)
	router.POST("/session", sessionWebHandler.Create)
	router.GET("/logout", sessionWebHandler.Destroy)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("server shutting down...")

	// Graceful shutdown with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped gracefully")

}

func authMiddleware(authService auth.Service, userService user.Service) gin.HandlerFunc {

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if !strings.Contains(authHeader, "Bearer") {
			response := helper.ApiResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		tokenString := ""
		arrayToken := strings.Split(authHeader, " ")
		if len(arrayToken) == 2 {
			tokenString = arrayToken[1]
		}

		token, err := authService.ValidateToken((tokenString))
		if err != nil {
			response := helper.ApiResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		claim, ok := token.Claims.(jwt.MapClaims)

		if !ok || !token.Valid {
			response := helper.ApiResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		userID := int(claim["user_id"].(float64))

		user, err := userService.GetUserByID(userID)
		if err != nil {
			response := helper.ApiResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("currentUser", user)
	}
}

func authAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		userIDsession := session.Get("userID")
		if userIDsession == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
	}

}

func loadTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.Renderer{}

	layouts, err := filepath.Glob(templatesDir + "/layouts/*.html")
	if err != nil {
		logger.Fatal("failed to load templates", zap.Error(err))
	}

	includes, err := filepath.Glob(templatesDir + "/**/*")
	if err != nil {
		logger.Fatal("failed to load templates", zap.Error(err))
	}

	// Generate our templates map from our layouts/ and includes/ directories
	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		r.AddFromFiles(filepath.Base(include), files...)
	}
	return r
}
