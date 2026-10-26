package main

import (
	"log"
	"siakad-mini/controllers"
	"siakad-mini/database"
	"siakad-mini/middleware"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	database.ConnectDB()
	database.MigrateDB()
	database.SeedDB()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())

	api := app.Group("/api/v1")

	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak percobaan login, silakan coba lagi nanti",
			})
		},
	})

	auth := api.Group("/auth")
	auth.Post("/login", loginLimiter, controllers.Login)
	auth.Get("/me", middleware.Protected(), controllers.GetMe)

	students := api.Group("/students", middleware.Protected())
	students.Get("/", middleware.AdminOnly(), controllers.GetStudents)
	students.Post("/", middleware.AdminOnly(), controllers.CreateStudent)
	students.Get("/:id", controllers.GetStudentDetail)
	students.Put("/:id", middleware.AdminOnly(), controllers.UpdateStudent)
	students.Delete("/:id", middleware.AdminOnly(), controllers.DeleteStudent)

	courses := api.Group("/courses", middleware.Protected())
	courses.Get("/", controllers.GetCourses)

	enrollments := api.Group("/enrollments", middleware.Protected(), middleware.MahasiswaOnly())
	enrollments.Post("/", controllers.CreateEnrollment)
	enrollments.Delete("/:id", controllers.DeleteEnrollment)

	log.Fatal(app.Listen(":3000"))
}
