package routes

import (
	"order-service/internal/handlers"
	"order-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/v1")
	//Payment group
	paymentRoutes := api.Group("payment")
	paymentRoutes.Post("/checkout", middleware.AuthUserMiddleware, handlers.PaymentCreate)
	paymentRoutes.Get("/success", handlers.PaymentSuccess)
	paymentRoutes.Get("/cancel", handlers.PaymentCancel)
	//Payment group employee
	paymentRoutes.Post("/:id/refund", middleware.AuthEmployeeMiddleware("admin", "manager"), handlers.PaymentRefund)

	// order group
	orderRoutes := api.Group("order")
	orderRoutes.Get("/", middleware.AuthUserMiddleware, handlers.OrderList)
	orderRoutes.Get("/:id", middleware.AuthUserMiddleware, handlers.OrderDetail)
	//Statistics group
	statisticsRoutes := api.Group("statistics")
	statisticsRoutes.Get("/order", middleware.AuthUserMiddleware, handlers.GetUserOrderStatistic)
	statisticsRoutes.Get("/payment", middleware.AuthEmployeeMiddleware("admin"), handlers.PaymentStatistics)

	// Statistics Employee group
	statisticsRoutes.Get("/sales-amount", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderPaymentStatistic)
	// statisticsRoutes.Get("/sales-general", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderGeneralStatistic)
	// statisticsRoutes.Get("/ranking-product", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderRankingProductStatistic)
}
