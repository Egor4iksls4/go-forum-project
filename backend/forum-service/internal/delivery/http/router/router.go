package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go-forum-project/forum-service/internal/delivery/http/handler"
	"go-forum-project/forum-service/internal/usecase"
)

func NewRouter(postUC usecase.PostUseCase, commentUC usecase.CommentUseCase, authMiddleware gin.HandlerFunc) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Refresh-Token"},
		ExposeHeaders:    []string{"Content-Length", "New-Access-Token", "New-Refresh-Token"}, // Добавлено
		AllowCredentials: true,
	}))

	postHandler := handler.NewPostHandler(postUC)
	commentHandler := handler.NewCommentHandler(commentUC)

	publicGroup := router.Group("/api")
	{
		publicGroup.GET("/posts", postHandler.GetAllPosts)

		postGroup := publicGroup.Group("/posts/:postId")
		{
			postGroup.GET("", postHandler.GetPostByID)
			postGroup.GET("/comments", commentHandler.GetCommentsByPostID)
		}
	}

	authGroup := router.Group("/api")
	authGroup.Use(authMiddleware)
	{
		authGroup.POST("/posts", postHandler.CreatePost)
		authGroup.PUT("/posts/:postId", postHandler.UpdatePost)
		authGroup.DELETE("/posts/:postId", postHandler.DeletePost)

		authGroup.POST("/posts/:postId/comments", commentHandler.CreateComment)
		authGroup.DELETE("/comments/:commentId", commentHandler.DeleteComment) // Единственный маршрут для удаления
	}

	return router
}
