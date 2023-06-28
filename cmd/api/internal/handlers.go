package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
}

func (h *Handlers) Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "holis",
	})

}
