package routes

import (
	"order-service/internal/handlers"
	"order-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/v1")
	// Định nghĩa các api của employee
	employeeGroup := api.Group("employee")
	//Payment group
	paymentRoutes := api.Group("payment")
	paymentRoutes.Post("/checkout", middleware.AuthUserMiddleware, handlers.PaymentCreate)
	paymentRoutes.Get("/success", handlers.PaymentSuccess)
	paymentRoutes.Get("/cancel", handlers.PaymentCancel)
	//Payment group employee
	employeePaymentRoutes := employeeGroup.Group("payment")
	employeePaymentRoutes.Post("/:id/refund", middleware.AuthEmployeeMiddleware("admin", "manager"), handlers.PaymentRefund)

	// order group
	orderRoutes := api.Group("order")
	orderRoutes.Get("/", middleware.AuthUserMiddleware, handlers.OrderList)
	// order group employee
	employeeOrderRoutes := employeeGroup.Group("order")
	employeeOrderRoutes.Get("/", middleware.AuthUserMiddleware, handlers.OrderList)
	//Statistics group
	// statisticsRoutes := api.Group("statistics")
	// Statistics Employee group
	EmployeeeStatisticsRoutes := employeeGroup.Group("statistics")
	EmployeeeStatisticsRoutes.Get("/sales-amount", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderPaymentStatistic)
	EmployeeeStatisticsRoutes.Get("/sales-general", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderGeneralStatistic)
	EmployeeeStatisticsRoutes.Get("/ranking-product", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderRankingProductStatistic)
	// EmployeeeStatisticsRoutes.Get("/ranking-user", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderPaymentStatistic)
	// EmployeeeStatisticsRoutes.Get("/ranking-order", middleware.AuthEmployeeMiddleware("admin"), handlers.GetOrderPaymentStatistic)

}
