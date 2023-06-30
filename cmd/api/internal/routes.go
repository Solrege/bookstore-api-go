package internal

import "github.com/gin-gonic/gin"

func InitRoutes(r *gin.Engine) {
	h := Handlers{}

	//rutas públicas

	r.GET("/", h.Index)
	r.POST("/register", h.RegisterHandler)
	r.POST("/login", h.LoginHandler)

	g := r.Group("/books")

	{
		g.GET("/:id", h.GetBookByIDHandler)
		g.GET("/:category", h.GetBooksByCategoryHandler)
		g.GET("/:author", h.GetBooksByAuthorHandler)
	}

	// rutas privadas

	a := r.Group("/admin")
	{
		a.POST("/books", h.AddNewBookHandler)
	}
}
