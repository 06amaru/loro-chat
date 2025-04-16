package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"server/controllers"
	"server/utils"

	"server/db"
	"server/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	// LOAD VAR IN LOCAL ENVIRONMENT
	_ = godotenv.Load(".env")
	utils.MySigningKey = []byte(os.Getenv("SIGNING_KEY"))
}

func main() {
	postgresRepo, err := db.NewPostgresRepository()
	if err != nil {
		log.Fatalf("Failure database when the server starts [%v]", err)
	}

	defer postgresRepo.Close()

	// Echo instance
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health-check", func(ctx echo.Context) error { return ctx.JSON(200, models.HealthCheck{Status: "UP"}) })

	authController := controllers.NewAuthController(postgresRepo)

	// curl -X POST -H 'Content-Type: application/json' -d '{"username":"jaoks", "password":"sdtc"}' localhost:8081/login
	e.POST("/login", authController.SignIn)

	chatController := controllers.NewChatController(postgresRepo)
	protected := e.Group("/api")

	protected.Use(utils.CustomMiddleware)

	// curl localhost:8081/api/chats --cookie "token=<YOUR_TOKEN>"
	protected.GET("/chats", chatController.GetChats)

	// curl "localhost:8081/api/:chatID/messages?limit=5&offset=0" --cookie "token=<YOUR_TOKEN>"
	protected.GET("/:chatID/messages", chatController.GetMessages)

	sockets := e.Group("/ws")
	sockets.Use(utils.CustomMiddleware)
	// websocat "ws://localhost:8081/ws/join?id=<CHAT_ID>" -H "Cookie: token=<YOUR_TOKEN>"
	sockets.GET("/join", chatController.JoinChat)

	// Start server
	e.Logger.Fatal(e.Start(":8081"))
}
