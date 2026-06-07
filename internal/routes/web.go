package routes

import (
	"backend/campaign"
	"backend/internal/handler/web"
	"backend/internal/middleware"
	"backend/transaction"
	"backend/user"

	"github.com/gin-gonic/gin"
)

// RegisterWeb routes (admin panel with HTML templates)
func RegisterWeb(
	router *gin.Engine,
	userService user.Service,
	campaignService campaign.Service,
	transactionService transaction.Service,
) {
	userH := handlerweb.NewUserHandler(userService)
	campaignH := handlerweb.NewCampaignHandler(campaignService, userService)
	transactionH := handlerweb.NewTransactionHandler(transactionService)
	sessionH := handlerweb.NewSessionHandler(userService)

	adminMW := middleware.AuthAdminMiddleware()

	// Users
	router.GET("/users", adminMW, userH.Index)
	router.GET("/users/new", adminMW, userH.New)
	router.POST("/users", adminMW, userH.Create)
	router.GET("/users/edit/:id", adminMW, userH.Edit)
	router.POST("/users/update/:id", adminMW, userH.Update)
	router.GET("/users/delete/:id", adminMW, userH.Delete)
	router.GET("/users/avatar/:id", adminMW, userH.NewAvatar)
	router.POST("/users/avatar/:id", adminMW, userH.CreateAvatar)

	// Campaigns
	router.GET("/campaigns", adminMW, campaignH.Index)
	router.GET("/campaigns/new", adminMW, campaignH.New)
	router.POST("/campaigns", adminMW, campaignH.Create)
	router.GET("/campaigns/image/:id", adminMW, campaignH.NewImage)
	router.POST("/campaigns/image/:id", adminMW, campaignH.CreateImage)
	router.GET("/campaigns/edit/:id", adminMW, campaignH.Edit)
	router.POST("/campaigns/update/:id", adminMW, campaignH.Update)
	router.GET("/campaigns/show/:id", adminMW, campaignH.Show)

	// Transactions
	router.GET("/transactions", adminMW, transactionH.Index)

	// Sessions (login/logout)
	router.GET("/login", sessionH.Index)
	router.POST("/session", sessionH.Create)
	router.GET("/logout", sessionH.Destroy)
}