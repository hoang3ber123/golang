package routes

import (
	"order-service/internal/handlers"
	"order-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/v1")
	// Định nghĩa các route
	//Payment group
	paymentRoutes := api.Group("payment")
	paymentRoutes.Post("/checkout", middleware.AuthUserMiddleware, handlers.PaymentCreate)
	paymentRoutes.Get("/success", handlers.PaymentSuccess)
	paymentRoutes.Get("/cancel", handlers.PaymentCancel)
	// order group
	orderRoutes := api.Group("order")
	orderRoutes.Get("/", middleware.AuthUserMiddleware, handlers.OrderList)
	//Statistics group
	statisticsRoutes := api.Group("statistics")
	statisticsRoutes.Get("/sales-amount", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderPaymentStatistic)
	statisticsRoutes.Get("/sales-general", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderGeneralStatistic)
	statisticsRoutes.Get("/ranking-product", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderRankingProductStatistic)
	// statisticsRoutes.Get("/ranking-user", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderPaymentStatistic)
	// statisticsRoutes.Get("/ranking-order", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderPaymentStatistic)

}
