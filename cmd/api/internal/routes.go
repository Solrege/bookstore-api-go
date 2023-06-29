package internal

import "github.com/gin-gonic/gin"

func InitRoutes(r *gin.Engine) {
	h := Handlers{}

	r.GET("/", h.Index)
	r.POST("/register", h.RegisterHandler)
	r.POST("/login", h.LoginHandler)
}
