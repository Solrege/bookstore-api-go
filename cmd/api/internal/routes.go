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
		g.GET("/:ID", h.GetBookByIDHandler)
		g.GET("/category/:category", h.GetBooksByCategoryHandler)
		g.GET("/author/:author", h.GetBooksByAuthorHandler)
		g.GET("/search", h.SearchBookHandler)
	}

	// rutas privadas de admin

	a := r.Group("/admin", JwtAuthMiddleware(), AdminAuthMiddleware())
	{
		a.POST("/books", h.AddNewBookHandler)
		a.DELETE("/books/:ID", h.DeleteBookHandler)
		a.PATCH("/books/:ID", h.UpdateBookHandler)
		// cancelar una orden

	}

	// rutas privadas de user
	u := r.Group("/user", JwtAuthMiddleware())
	{
		u.GET("/address", h.GetAddressHandler)
		u.POST("/address", h.AddAddressHandler)
		u.PATCH("/address", h.UpdateAddressHandler)

		// agregar al carrito
		// confirmar compra
		u.POST("/order")

		// ir a mercado pago
		// ver historial de compra

	}

	o := r.Group("/order", JwtAuthMiddleware())
	{
		// agregar al carrito
		// confirmar compra
		o.POST("/", h.CreateOrderHandler)

		// ir a mercado pago
		// ver historial de compra

	}
}
