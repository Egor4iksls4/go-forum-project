package handler

import (
	"github.com/gin-gonic/gin"
	"go-forum-project/forum-service/internal/usecase"
	"log"
	"net/http"
	"strconv"
)

type CommentHandler struct {
	commentUC usecase.CommentUseCase
}

func NewCommentHandler(commentUC usecase.CommentUseCase) *CommentHandler {
	return &CommentHandler{commentUC: commentUC}
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("postId"))
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		log.Println("Username not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	log.Printf("Creating comment for post %d by user %s", postID, username.(string))
	err = h.commentUC.Create(c.Request.Context(), postID, req.Content, username.(string))
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "failed to create comment",
		})
		return
	}

	log.Println("Comment created successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "comment created successfully"})
}

func (h *CommentHandler) GetCommentsByPostID(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("postId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	comments, err := h.commentUC.GetByPostID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
	})
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.Atoi(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = h.commentUC.DeleteComment(c.Request.Context(), commentID, username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted successfully"})
}
