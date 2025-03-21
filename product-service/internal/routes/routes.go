package routes

import (
	"product-service/internal/handlers"
	"product-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/v1")

	//Category group
	categoryRoutes := api.Group("category")
	categoryRoutes.Post("/", middleware.AuthEmployeeMiddleware("admin"), handlers.CategoryCreate)
	categoryRoutes.Get("/", handlers.CategoryList)
	categoryRoutes.Delete("/", middleware.AuthEmployeeMiddleware("admin"), handlers.CategoryDelete)
	categoryRoutes.Patch("/:id", middleware.AuthEmployeeMiddleware("admin"), handlers.CategoryUpdate)
	categoryRoutes.Get("/:slug", handlers.CategoryDetail)

	// Product group
	productRoutes := api.Group("product")
	productRoutes.Post("/", middleware.AuthEmployeeMiddleware("admin"), handlers.ProductCreate)
	productRoutes.Get("/:id/download", middleware.AuthUserMiddleware, handlers.ProductDownload)
	productRoutes.Get("/order/bought", middleware.AuthUserMiddleware, handlers.ProductFromOrder)
	productRoutes.Patch("/:id", middleware.AuthEmployeeMiddleware("admin"), handlers.ProductUpdate)
	productRoutes.Delete("/", middleware.AuthEmployeeMiddleware("admin"), handlers.ProductDelete)
	productRoutes.Get("/", handlers.ProductList)
	productRoutes.Get("/:slug", handlers.ProductDetail)
	// get role detail
	// create role
	// delete role
	// patch role
	// get role list
}
