package internal

import (
	"bookstore-api/internal/business"
	"bookstore-api/internal/platform"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := TokenValid(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized, token not valid")
			fmt.Println(err.Error())
			c.Abort()
			return
		}

		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id, _ := c.Get("user_id")

		db := platform.DbConnection()
		var user business.User

		result := db.Where("ID = ?", user_id).Find(&user)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Something went wrong with the ID",
			})

			return
		}

		if user.Role != "admin" {
			c.String(http.StatusUnauthorized, "Unauthorized, not an admin")
			c.Abort()

			return
		}

		c.Next()
	}
}
