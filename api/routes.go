package api

import (
	"Backend/configs"
	"Backend/internal/handlers/aspirations"
	"Backend/internal/handlers/auth"
	"Backend/internal/handlers/candidate"
	"Backend/internal/handlers/event"
	"Backend/internal/handlers/news"
	"Backend/internal/handlers/permission"
	"Backend/internal/handlers/role"
	"Backend/internal/handlers/user"
	"Backend/internal/handlers/version"
	"Backend/internal/handlers/vote"
	"Backend/internal/middleware"
	"Backend/internal/services"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes() *gin.Engine {
	// Set Gin to release mode for better performance
	gin.SetMode(gin.ReleaseMode)
	
	// Create a new engine with default configuration
	r := gin.New()
	
	// Add recovery middleware to recover from panics
	r.Use(gin.Recovery())
	
	// Use custom logger for better performance under high load
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/public", "/api/v1/health"}, // Skip logging for static files and health checks
	}))

	// Set maximum multipart memory limit to 10MB
	r.MaxMultipartMemory = 10 << 20 // 10 MB

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{
			"https://computing.president.ac.id", 
			"https://staging.computing.president.ac.id", 
			"https://compsci.president.ac.id", 
			"https://staging.compsci.president.ac.id", 
			"http://localhost:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Increase rate limits for high traffic
	maxTokens := 2000 // Doubled from 1000 to handle 100+ concurrent users
	refillInterval := time.Minute
	r.Use(middleware.RateLimiterMiddleware(maxTokens, refillInterval, "general"))
	
	// Add a health check endpoint that bypasses rate limiting
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	r.Static("/public", "./public")

	authService := services.NewAuthService()
	userService := services.NewUserService()
	eventService := services.NewEventService()
	newsService := services.NewNewsService()
	roleService := services.NewRoleService()
	permissionService := services.NewPermissionService()
	aspirationsService := services.NewAspirationService()
	candidateService := services.NewCandidateService()
	voteService := services.NewVoteService()
	AWSService, _ := services.NewAWSService()
	R2Service, _ := services.NewR2Service()
	// Get email service configuration
	config := configs.LoadConfig()

	// Initialize email service based on configuration
	var EmailService services.EmailService

	// Check if we should use SMTP or SendGrid
	if config.UseSmtp {
		// Use SMTP service
		smtpHost := config.SMTPHost
		if smtpHost == "" {
			smtpHost = "smtp.gmail.com"
			log.Println("Using fallback SMTP host: smtp.gmail.com")
		}

		smtpPort := config.SMTPPort
		if smtpPort == "" {
			smtpPort = "587"
			log.Println("Using fallback SMTP port: 587")
		}

		smtpUsername := config.SMTPUsername
		if smtpUsername == "" {
			log.Println("WARNING: SMTP username not found in environment variables")
		}

		smtpPassword := config.SMTPPassword
		if smtpPassword == "" {
			log.Println("WARNING: SMTP password not found in environment variables")
		}

		senderEmail := config.SenderEmail
		if senderEmail == "" {
			log.Println("WARNING: SMTP sender email not found in environment variables")
		}

		EmailService = services.NewTestMailService(
			smtpHost,
			smtpPort,
			smtpUsername,
			smtpPassword,
			senderEmail,
		)
		log.Println("Using SMTP email service")
	} else {

		brevoAPIKey := config.BrevoAPIKey
		if brevoAPIKey == "" {
			log.Println("Warning: Brevo API key not provided, email functionality will be limited")
		}

		brevoSenderEmail := config.BrevoSenderEmail
		if brevoSenderEmail == "" {
			brevoSenderEmail = "noreply@pufacomputing.ac.id" // Default sender email
			log.Println("Using default sender email:", brevoSenderEmail)
		}

		brevoSenderName := config.BrevoSenderName
		if brevoSenderName == "" {
			brevoSenderName = "PUFA Computer Science"
			log.Println("Using default sender name:", brevoSenderName)
		}

		EmailService = services.NewBrevoService(
			brevoAPIKey,
			brevoSenderEmail,
			brevoSenderName,
		)
		log.Println("Using Brevo email service")
	}
	VersionService := services.NewVersionService(configs.LoadConfig().GithubAccessToken)

	eventStatusUpdater := services.NewEventStatusUpdater(eventService)
	go eventStatusUpdater.Run()

	versionUpdater := services.NewVersionUpdater(VersionService)
	go versionUpdater.Run()

	authHandlers := auth.NewAuthHandlers(authService, permissionService, EmailService, userService)
	userHandlers := user.NewUserHandlers(userService, permissionService, AWSService, R2Service)
	eventHandlers := event.NewEventHandlers(eventService, permissionService, AWSService, R2Service)
	newsHandlers := news.NewNewsHandler(newsService, permissionService, AWSService, R2Service)
	roleHandlers := role.NewRoleHandler(roleService, userService, permissionService)
	permissionHandlers := permission.NewPermissionHandler(permissionService)
	aspirationHandlers := aspirations.NewAspirationHandlers(aspirationsService, permissionService)
	candidateHandlers := candidate.NewCandidateHandler(candidateService, permissionService, AWSService, R2Service)
	voteHandlers := vote.NewVoteHandler(voteService, permissionService)
	versionHandlers := version.NewVersionHandlers(VersionService)

	api := r.Group("/api/v1")

	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/register", authHandlers.RegisterUser)
		authRoutes.POST("/login", middleware.RateLimiterMiddleware(100, time.Minute, "login"), authHandlers.Login)
		authRoutes.POST("/logout", authHandlers.Logout)
		authRoutes.POST("/refresh-token", middleware.TokenMiddleware(), authHandlers.RefreshToken)
		authRoutes.GET("/verify-email", authHandlers.VerifyEmail)
		authRoutes.POST("/forgot-password/request", authHandlers.RequestPasswordReset)
		authRoutes.POST("/forgot-password", authHandlers.ResetPassword)
	}

	userRoutes := api.Group("/user")
	{
		userRoutes.Use(middleware.TokenMiddleware())
		userRoutes.GET("/:userID", userHandlers.GetUserByID)
		userRoutes.PUT("/edit", userHandlers.EditUser)
		userRoutes.DELETE("/delete", userHandlers.DeleteUser)
		userRoutes.PUT("/change-password", userHandlers.ChangePassword)
		userRoutes.POST("/upload-profile-picture", userHandlers.UploadProfilePicture)
		userRoutes.POST("/upload-student-id", userHandlers.UploadStudentID)
		userRoutes.PUT("/:userID/update-user", userHandlers.AdminUpdateRoleAndStudentIDVerified)
		userRoutes.POST("/2fa/enable", userHandlers.EnableTwoFA)
		userRoutes.POST("/2fa/verify", userHandlers.VerifyTwoFA)
		userRoutes.POST("/2fa/toggle", userHandlers.ToggleTwoFA)

		// ListEventsRegisteredByUser
		userRoutes.GET("/registered-events", eventHandlers.ListEventsRegisteredByUser)
	}

	// Admin routes for user management
	adminRoutes := api.Group("/admin")
	{
		adminRoutes.Use(middleware.TokenMiddleware())
		adminRoutes.GET("/users", userHandlers.ListUsers)              // original endpoint for admin to list all users
		adminRoutes.GET("/users/basic", userHandlers.GetAllUsersBasic) // new endpoint that avoids NULL issues
	}

	eventRoutes := api.Group("/event")
	{
		eventRoutes.GET("/:eventID", eventHandlers.GetEventBySlug)
		eventRoutes.GET("/", eventHandlers.ListEvents)
		eventRoutes.GET("/:eventID/total-participant", eventHandlers.TotalRegisteredUsers)
		eventRoutes.Use(middleware.TokenMiddleware())
		eventRoutes.POST("/create", eventHandlers.CreateEvent)
		eventRoutes.PATCH("/:eventID/edit", eventHandlers.EditEvent)
		eventRoutes.DELETE("/:eventID/delete", eventHandlers.DeleteEvent)
		eventRoutes.POST("/:eventID/register", eventHandlers.RegisterForEvent)
		eventRoutes.GET("/:eventID/registered-users", eventHandlers.ListRegisteredUsers)
	}

	newsRoutes := api.Group("/news")
	{
		newsRoutes.GET("/", newsHandlers.ListNews)
		newsRoutes.GET("/:newsID", newsHandlers.GetNewsBySlug)
		newsRoutes.Use(middleware.TokenMiddleware())
		newsRoutes.POST("/create", newsHandlers.CreateNews)
		newsRoutes.PUT("/:newsID/edit", newsHandlers.EditNews)
		newsRoutes.DELETE("/:newsID/delete", newsHandlers.DeleteNews)
		newsRoutes.POST("/:newsID/like", newsHandlers.LikeNews)
	}

	roleRoutes := api.Group("/roles")
	{
		roleRoutes.Use(middleware.TokenMiddleware())
		roleRoutes.GET("/", roleHandlers.ListRoles)
		roleRoutes.POST("/create", roleHandlers.CreateRole)
		roleRoutes.GET("/:roleID", roleHandlers.GetRoleByID)
		roleRoutes.PUT("/:roleID/edit", roleHandlers.EditRole)
		roleRoutes.DELETE("/:roleID/delete", roleHandlers.DeleteRole)
		roleRoutes.POST("/:roleID/assign/:userID", roleHandlers.AssignRoleToUser)
	}
	permissionRoutes := api.Group("/permissions")
	{
		permissionRoutes.Use(middleware.TokenMiddleware())
		permissionRoutes.GET("/list", permissionHandlers.ListPermissions)
		permissionRoutes.POST("/assign/:roleID", permissionHandlers.AssignPermissionToRole)

	}

	aspirationRoutes := api.Group("/aspirations")
	{
		aspirationRoutes.GET("/", aspirationHandlers.GetAspirations)
		aspirationRoutes.GET("/:id", aspirationHandlers.GetAspirationByID)
		aspirationRoutes.Use(middleware.TokenMiddleware())
		aspirationRoutes.POST("/create", aspirationHandlers.CreateAspiration)
		aspirationRoutes.PATCH("/:id/close", aspirationHandlers.CloseAspiration)
		aspirationRoutes.DELETE("/:id/delete", aspirationHandlers.DeleteAspiration)
		aspirationRoutes.POST("/:id/upvote", aspirationHandlers.UpvoteAspiration)
		aspirationRoutes.GET("/:id/get_upvotes", aspirationHandlers.GetUpvotesByAspirationID)
		aspirationRoutes.POST("/:id/admin_reply", aspirationHandlers.AddAdminReply)
	}

	versionRoutes := api.Group("/version")
	{
		versionRoutes.GET("/", versionHandlers.GetVersion)
		versionRoutes.GET("/changelog", versionHandlers.GetChangelog)
	}

	// Authenticated candidate routes (registered first to avoid route conflicts)
	candidateAuthRoutes := api.Group("/candidates")
	{
		candidateAuthRoutes.Use(middleware.TokenMiddleware())
		candidateAuthRoutes.GET("/my-major", candidateHandlers.GetCandidatesForMyMajor)
		candidateAuthRoutes.POST("/create", candidateHandlers.CreateCandidate)
		candidateAuthRoutes.PUT("/:candidateID/edit", candidateHandlers.UpdateCandidate)
		candidateAuthRoutes.DELETE("/:candidateID/delete", candidateHandlers.DeleteCandidate)
	}
	// Public candidate routes
	candidateRoutes := api.Group("/candidates")
	{
		candidateRoutes.GET("/", candidateHandlers.GetAllCandidates)
		candidateRoutes.GET("/:candidateID", candidateHandlers.GetCandidateByID)
	}

	voteRoutes := api.Group("/votes")
	{
		voteRoutes.Use(middleware.TokenMiddleware())
		voteRoutes.POST("/cast", voteHandlers.CastVote)
		voteRoutes.GET("/status", voteHandlers.GetVoteStatus)
		voteRoutes.GET("/my-vote", voteHandlers.GetMyVote)
		voteRoutes.GET("/can-vote", voteHandlers.CheckCanVote)
		voteRoutes.GET("/", voteHandlers.ListVotes)
		voteRoutes.GET("/candidate/:candidateID", voteHandlers.GetVotesByCandidateID)
		voteRoutes.DELETE("/:voteID/delete", voteHandlers.DeleteVote)
	}
	// Public vote routes (no authentication required)
	votePublicRoutes := api.Group("/votes")
	{
		votePublicRoutes.GET("/candidate/:candidateID/count", voteHandlers.GetVoteCountByCandidate)
	}

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
