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
	categoryRoutes.Get("/all", handlers.CategoryAllList)
	categoryRoutes.Delete("/", middleware.AuthEmployeeMiddleware("admin"), handlers.CategoryDelete)
	categoryRoutes.Get("/recommend", middleware.AuthUserMiddleware, handlers.CategoryRecommend)
	categoryRoutes.Patch("/:id", middleware.AuthEmployeeMiddleware("admin"), handlers.CategoryUpdate)
	categoryRoutes.Get("/:slug", handlers.CategoryDetail)

	// Product group
	productRoutes := api.Group("product")
	productRoutes.Post("/", middleware.AuthEmployeeMiddleware("admin"), handlers.ProductCreate)
	productRoutes.Get("/:id/download", middleware.AuthUserMiddleware, handlers.ProductDownload)
	productRoutes.Get("/order/bought", middleware.AuthUserMiddleware, handlers.ProductFromOrder)
	productRoutes.Patch("/:id", middleware.AuthEmployeeMiddleware("admin"), handlers.ProductUpdate)
	productRoutes.Delete("/", middleware.AuthEmployeeMiddleware("admin"), handlers.ProductDelete)
	productRoutes.Get("/", middleware.DefaultMiddleware, handlers.ProductList)
	productRoutes.Get("/recommend", middleware.AuthUserMiddleware, handlers.ProductRecommend)

	productRoutes.Get("/:slug", middleware.DefaultMiddleware, handlers.ProductDetail)

	// Cart Group
	cartRoutes := api.Group("cart")
	cartRoutes.Post("/product", middleware.AuthUserMiddleware, handlers.CartProductAdd)
	cartRoutes.Delete("/product", middleware.AuthUserMiddleware, handlers.CartProductRemove)
	cartRoutes.Get("/product/detail", middleware.AuthUserMiddleware, handlers.CartProductDetail)

	// UploadFile to vstorage
	vstorage := api.Group("vstorage")
	vstorage.Post("/upload", middleware.AuthEmployeeMiddleware("admin", "manager"), handlers.UploadMedia)

	// get role detail
	// create role
	// delete role
	// patch role
	// get role list
}
