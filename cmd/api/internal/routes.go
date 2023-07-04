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
		g.GET("/search", h.SearchBookHandler)
	}

	// rutas privadas de admin

	a := r.Group("/admin")
	{
		a.POST("/books", h.AddNewBookHandler)
		a.DELETE("/books/:ID", h.DeleteBookHandler)
		a.PATCH("/books/:ID", h.UpdateBookHandler)
		// cancelar una orden

	}

	// rutas privadas de user
	//u := r.Group("/user")
	{
		// agregar al carrito
		// hacer compra
		// completar datos de direccion y pago?
	}
}
