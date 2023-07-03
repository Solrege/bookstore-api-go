package internal

import "github.com/gin-gonic/gin"

func InitRoutes(r *gin.Engine) {
	h := Handlers{}

	//rutas públicas

	r.GET("/", h.Index)
	r.POST("/register", h.RegisterHandler) // falta
	r.POST("/login", h.LoginHandler)       // falta

	g := r.Group("/books")
	{
		g.GET("/:ID", h.GetBookByIDHandler)
		g.GET("/category/:category", h.GetBooksByCategoryHandler)
		g.GET("/author/:author", h.GetBooksByAuthorHandler)
	}

	// rutas privadas

	a := r.Group("/admin")
	{
		a.POST("/books", h.AddNewBookHandler)
	}
}
