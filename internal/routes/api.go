package routes

import (
	"backend/auth"
	"backend/campaign"
	"backend/internal/handler/api"
	"backend/internal/middleware"
	"backend/transaction"
	"backend/user"

	"github.com/gin-gonic/gin"
)

// RegisterAPI routes under /api/v1
func RegisterAPI(
	router *gin.RouterGroup,
	userService user.Service,
	authService auth.Service,
	campaignService campaign.Service,
	transactionService transaction.Service,
) {
	userH := handlerapi.NewUserHandler(userService, authService)
	campaignH := handlerapi.NewCampaignHandler(campaignService)
	transactionH := handlerapi.NewTransactionHandler(transactionService)

	authMW := middleware.AuthMiddleware(authService, userService)

	// Users
	router.POST("/users", userH.RegisterUser)
	router.POST("/sessions", userH.LoginUser)
	router.POST("/email_checkers", userH.CheckEmailAvailability)
	router.POST("/avatars", authMW, userH.UploadAvatar)
	router.GET("/users/fetch", authMW, userH.FetchUser)

	// Campaigns
	router.GET("/campaigns", campaignH.GetCampaigns)
	router.GET("/campaigns/:id", campaignH.GetCampaign)
	router.POST("/campaigns", authMW, campaignH.CreateCampaign)
	router.PUT("/campaigns/:id", authMW, campaignH.UpdateCampaign)
	router.POST("/campaign-images", authMW, campaignH.UploadImage)

	// Transactions
	router.GET("/campaigns/:id/transactions", authMW, transactionH.GetCampaignTransactions)
	router.GET("/transactions", authMW, transactionH.GetUserTransactions)
	router.POST("/transactions", authMW, transactionH.CreateTransaction)
	router.POST("/transactions/notification", transactionH.GetNotification)
}