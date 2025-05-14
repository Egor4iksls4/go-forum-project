package router

import (
	"github.com/gin-gonic/gin"
	"go-forum-project/forum-service/internal/delivery/http/handler"
	"go-forum-project/forum-service/internal/usecase"
)

func NewRouter(postUC *usecase.PostUseCase, authMiddleware gin.HandlerFunc) *gin.Engine {
	router := gin.Default()

	postHandler := handler.NewPostHandler(*postUC)

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
