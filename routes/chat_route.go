package routes

import (
	"context"
	"net/http"
	"tbibi_back_end_go/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)



func SetupChatRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	// Endpoint to search users
	r.GET("/api/v1/search/:username", func(c *gin.Context) {
		services.SearchUsers(c, pool)
	})

	// Endpoint to retrieve messages for a specific chat
	r.GET("/api/v1/messages/:chatId", func(c *gin.Context) {
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acquire a database connection"})
			return
		}
		defer conn.Release()
		services.GetMessagesForChat(conn.Conn(), c)
	})

	// Endpoint to create or find an existing chat between two users
	r.GET("/api/findOrCreateChat", func(c *gin.Context) {
		conn, err := pool.Acquire(context.Background())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acquire a database connection"})
				return
			}
			defer conn.Release()
			services.FindOrCreateChatWithUser(conn.Conn(), c)
		})
	


	r.GET("/api/v1/chats", func(c *gin.Context) {
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acquire a database connection"})
			return
		}
		defer conn.Release()
		services.ListChatsForUser(conn.Conn(), c)
	})

	r.POST("/api/v1/SendMessage", func(c *gin.Context) {
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acquire a database connection"})
			return
		}
		defer conn.Release()
		services.SendMessage(conn.Conn(), c)
	})

}