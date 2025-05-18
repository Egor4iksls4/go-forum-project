package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go-forum-project/forum-service/internal/delivery/http/handler"
	"go-forum-project/forum-service/internal/usecase"
)

func NewRouter(postUC usecase.PostUseCase, authMiddleware gin.HandlerFunc) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Refresh-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	postHandler := handler.NewPostHandler(postUC)

	authGroup := router.Group("/api")
	authGroup.Use(authMiddleware)
	{
		authGroup.POST("/posts", postHandler.CreatePost)
		authGroup.PUT("/posts/:id", postHandler.UpdatePost)
		authGroup.DELETE("/posts/:id", postHandler.DeletePost)
	}

	publicGroup := router.Group("/api")
	{
		publicGroup.GET("/posts", postHandler.GetAllPosts)
	}

	return router
}
