package routes

import (
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/v1")

	//Auth group
	authRoutes := api.Group("auth")
	authRoutes.Post("/signup", handlers.SignUp)
	authRoutes.Post("/verify-email/:token", handlers.VerifyEmail)
	authRoutes.Post("/login", handlers.Login)
	authRoutes.Post("/logout", middleware.JWTAuthMiddleware, handlers.Logout)
	authRoutes.Post("/login/employee", handlers.EmployeeLogin)
	authRoutes.Post("/logout/employee", middleware.JWTAuthEmployeeMiddleware, handlers.EmployeeLogout)

	// User group
	userRoutes := api.Group("user")
	userRoutes.Get("/", middleware.JWTAuthEmployeeMiddleware, middleware.RestrictRoleMiddlware("admin"), handlers.UserList)
	userRoutes.Get("/detail", middleware.JWTAuthMiddleware, handlers.UserDetail)
	// userRoutes.Get("/:id", middleware.JWTAuthEmployeeMiddleware, middleware.RestrictRoleMiddlware("admin"), handlers.UserAdminDetail)
	employeeRoutes := api.Group("employee")
	employeeRoutes.Use(middleware.JWTAuthEmployeeMiddleware) // Áp dụng việc đăng nhập cho tất cả api thuộc employee
	employeeRoutes.Post("/", middleware.RestrictRoleMiddlware("admin"), handlers.EmployeeCreate)
	employeeRoutes.Get("/", middleware.RestrictRoleMiddlware("admin"), handlers.EmployeeList)
	employeeRoutes.Get("/detail", handlers.EmployeeDetail)

	// login
	// logout
	// decentralize
	// employeeRoutes.Post("/decentralize", middleware.JWTAuthMiddleware, handlers.Decentralize)
	// create manager
	// view user

	// check role
	roleRoutes := api.Group("role")
	roleRoutes.Use(middleware.JWTAuthEmployeeMiddleware) // Áp dụng việc đăng nhập cho tất cả api thuộc employee
	roleRoutes.Post("/", middleware.RestrictRoleMiddlware("admin"), handlers.RoleCreate)
	roleRoutes.Get("/", handlers.RoleList)
	roleRoutes.Delete("/", middleware.RestrictRoleMiddlware("admin"), handlers.RoleDelete)
	roleRoutes.Patch("/:id", middleware.RestrictRoleMiddlware("admin"), handlers.RoleUpdate)
	roleRoutes.Get("/:slug", handlers.RoleDetail)
	// get role detail
	// create role
	// delete role
	// patch role
	// get role list
}
