package routes

import (
	"golang-backend/controller"
	"golang-backend/middleware"
	"golang-backend/utils"

	_ "golang-backend/docs"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(app *gin.Engine, userCtrl *controller.UserController, roleCtrl *controller.RoleController, examCtrl *controller.ExamController, sessionCtrl *controller.SessionController, questionCtrl *controller.QuestionController, candidateCtrl *controller.CandidateController, assignmentCtrl *controller.AssignmentController, attemptCtrl *controller.AttemptController) {
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := app.Group("/api")
	api.GET("/ping", func(c *gin.Context) {
		utils.SuccessResponse(c, "pong", nil)
	})
	api.POST("/register", userCtrl.Register)
	api.POST("/login", userCtrl.Login)
	api.POST("/verify-email", userCtrl.VerifyEmail)
	api.POST("/forgot-password", userCtrl.ForgotPassword)
	api.POST("/reset-password", userCtrl.ResetPassword)
	api.POST("/resend-verification", userCtrl.ResendVerificationCode)
	api.POST("/resend-reset-code", userCtrl.ResendResetPasswordCode)

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/me", userCtrl.Me)
	protected.GET("/users", userCtrl.GetUsers)
	protected.POST("/2fa/setup", userCtrl.Setup2FA)
	protected.POST("/2fa/verify", userCtrl.Verify2FA)

	admin := protected.Group("/admin")
	admin.Use(middleware.RoleAuthMiddleware("admin"))
	admin.POST("/roles", roleCtrl.CreateRole)
	admin.POST("/permissions", roleCtrl.CreatePermission)
	admin.POST("/assign-role", roleCtrl.AssignRoleToUser)
	admin.POST("/assign-permission", roleCtrl.AssignPermissionToRole)

	apiV1 := app.Group("/api/v1")
	apiV1.GET("/ping", func(c *gin.Context) {
		utils.SuccessResponse(c, "pong", nil)
	})
	apiV1.POST("/auth/candidate/login", candidateCtrl.Login)

	adminV1 := apiV1.Group("/admin")
	adminV1.Use(middleware.AuthMiddleware(), middleware.RoleAuthMiddleware("admin"))
	adminV1.GET("/exams", examCtrl.List)
	adminV1.POST("/exams", examCtrl.Create)
	adminV1.GET("/exams/:id", examCtrl.Get)
	adminV1.PUT("/exams/:id", examCtrl.Update)
	adminV1.DELETE("/exams/:id", examCtrl.Archive)
	adminV1.GET("/sessions", sessionCtrl.List)
	adminV1.POST("/sessions", sessionCtrl.Create)
	adminV1.GET("/sessions/:id", sessionCtrl.Get)
	adminV1.PUT("/sessions/:id", sessionCtrl.Update)
	adminV1.DELETE("/sessions/:id", sessionCtrl.Archive)
	adminV1.GET("/questions", questionCtrl.List)
	adminV1.POST("/questions", questionCtrl.Create)
	adminV1.GET("/questions/:id", questionCtrl.Get)
	adminV1.PUT("/questions/:id", questionCtrl.Update)
	adminV1.DELETE("/questions/:id", questionCtrl.Archive)
	adminV1.GET("/candidates", candidateCtrl.List)
	adminV1.POST("/candidates", candidateCtrl.Create)
	adminV1.GET("/candidates/:id", candidateCtrl.Get)
	adminV1.PUT("/candidates/:id", candidateCtrl.Update)
	adminV1.DELETE("/candidates/:id", candidateCtrl.Archive)
	adminV1.GET("/assignments", assignmentCtrl.List)
	adminV1.POST("/assignments", assignmentCtrl.Create)
	adminV1.GET("/assignments/:id", assignmentCtrl.Get)
	adminV1.DELETE("/assignments/:id", assignmentCtrl.Archive)

	candidateV1 := apiV1.Group("/candidate")
	candidateV1.Use(middleware.CandidateAuthMiddleware())
	candidateV1.GET("/me", candidateCtrl.Me)
	candidateV1.GET("/exams", attemptCtrl.ListSchedules)
	candidateV1.POST("/exams/:exam_id/attempts", attemptCtrl.Start)
	candidateV1.GET("/attempts/:id", attemptCtrl.Get)
	candidateV1.PUT("/attempts/:id/answers/:question_id", attemptCtrl.Answer)
	candidateV1.POST("/attempts/:id/submit", attemptCtrl.Submit)
	candidateV1.GET("/attempts/:id/result", attemptCtrl.Result)
}
