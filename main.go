package main

import (
	"flag"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"golang-backend/config"
	"golang-backend/controller"
	"golang-backend/middleware"
	"golang-backend/migrations"
	"golang-backend/repository"
	"golang-backend/routes"
	"golang-backend/service"
	"golang-backend/utils"

	"github.com/gin-gonic/gin"
)

// @title           Golang Backend API
// @version         1.0
// @description     This is a sample server for Golang Backend.
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Load environment variables
	config.LoadEnv()

	// Setup logging with rotation (lumberjack)
	logFile := &lumberjack.Logger{
		Filename:   "logs/server.log",
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30,   // days
		Compress:   true, // disabled by default
	}

	// Set gin writer to both file and console
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(gin.DefaultWriter)

	// Setup slog for structured logging (JSON)
	logger := slog.New(slog.NewJSONHandler(gin.DefaultWriter, nil))
	slog.SetDefault(logger)

	// 2. Initialize database
	db := config.InitDB()

	// 3. Run Migrations (Optional via flag)
	migrate := flag.Bool("migrate", false, "Run database migrations")
	seed := flag.Bool("seed", false, "Run database seeder")
	seedDemo := flag.Bool("seed-demo", false, "Seed demo exam data and exit")
	flag.Parse()

	if *migrate {
		migrations.RunMigrations(db)
	}

	// 4. Initialize repositories
	userRepo := repository.NewUserRepository(db)
	examRepo := repository.NewExamRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	candidateRepo := repository.NewCandidateRepository(db)
	assignmentRepo := repository.NewAssignmentRepository(db)
	attemptRepo := repository.NewAttemptRepository(db)

	// 4. Initialize services
	userService := service.NewUserService(userRepo)
	examService := service.NewExamService(examRepo)
	sessionService := service.NewSessionService(sessionRepo)
	questionService := service.NewQuestionService(questionRepo)
	candidateService := service.NewCandidateService(candidateRepo)
	assignmentService := service.NewAssignmentService(assignmentRepo)
	attemptService := service.NewAttemptService(attemptRepo)
	go runAttemptExpiryWorker(attemptService, config.AppConfig.AttemptExpiryInterval)

	// 5. Initialize controllers
	userCtrl := controller.NewUserController(userService)
	examCtrl := controller.NewExamController(examService)
	sessionCtrl := controller.NewSessionController(sessionService)
	questionCtrl := controller.NewQuestionController(questionService)
	candidateCtrl := controller.NewCandidateController(candidateService)
	assignmentCtrl := controller.NewAssignmentController(assignmentService)
	attemptCtrl := controller.NewAttemptController(attemptService)
	roleCtrl := controller.NewRoleController(db)

	// Run Seeder
	if *seed {
		utils.SeedRolesAndPermissions(db)

		utils.SeedUsers(db)
	}
	if *seedDemo {
		if err := utils.SeedDemoExam(db); err != nil {
			log.Fatalf("Failed to seed demo exam: %v", err)
		}
		return
	}

	// 6. Initialize Gin router
	app := gin.New()
	app.Use(gin.Recovery())

	// 7. Apply global middleware
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.CORSMiddleware(config.AppConfig.CORSAllowedOrigins))
	app.Use(middleware.ErrorHandlerMiddleware())
	app.Use(middleware.RateLimiterMiddleware())

	// 8. Setup routes
	routes.SetupRoutes(app, userCtrl, roleCtrl, examCtrl, sessionCtrl, questionCtrl, candidateCtrl, assignmentCtrl, attemptCtrl)

	// 9. Liveness and readiness endpoints
	app.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, "API is running", gin.H{
			"status":     "ok",
			"request_id": middleware.RequestID(c),
		})
	})
	app.GET("/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			utils.ErrorResponse(c, "Database connection failed", http.StatusServiceUnavailable, gin.H{
				"status":     "not_ready",
				"request_id": middleware.RequestID(c),
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			utils.ErrorResponse(c, "Database ping failed", http.StatusServiceUnavailable, gin.H{
				"status":     "not_ready",
				"request_id": middleware.RequestID(c),
			})
			return
		}

		utils.SuccessResponse(c, "API is ready", gin.H{
			"status":     "ready",
			"request_id": middleware.RequestID(c),
		})
	})

	// 10. Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func runAttemptExpiryWorker(attemptService service.AttemptService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for now := range ticker.C {
		expiredCount, err := attemptService.ExpireDueAttempts(now.UTC())
		if err != nil {
			slog.Error("attempt_expiry_failed", "error", err)
			continue
		}
		if expiredCount > 0 {
			slog.Info("attempts_expired", "count", expiredCount)
		}
	}
}
