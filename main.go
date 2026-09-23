package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
)

func main() {

	config.LoadEnv()
	logger := config.NewLogger()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	roleRepository := repository.NewRoleRepository(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository, permissions)

	pr := repository.NewPrestasiRepository(pool)
	ps := service.NewPrestasiService(pr)

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", 32))
		os.Exit(1)
	}

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)

	authService := service.NewAuthService(
		userRepository, tokenRepository, jwtManager,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)

	deps := route.Dependencies{
		Pool:            pool,
		JWT:             jwtManager,
		Permissions:     permissions,
		UserService:     studentService,
		PrestasiService: ps,
		AuthService:     authService,
	}
	app := config.NewApp(logger, deps, config.GetEnv("ALLOWED_ORIGINS", ""))

	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("server berjalan", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi",
			slog.String("error", err.Error()))
	}

	logger.Info("server berhenti dengan rapi")
}
