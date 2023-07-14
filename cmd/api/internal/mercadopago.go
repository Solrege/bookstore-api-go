package internal

import (
	"bookstore-api/internal/business"
	"bookstore-api/internal/platform"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/eduardo-mior/mercadopago-sdk-go"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) CreatePayment(c *gin.Context) {
	ID := c.Param("ID")
	id, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the ID",
		})

		return
	}

	var order business.Order

	db := platform.DbConnection()
	result := db.Where("ID = ?", id).Preload("User").Find(&order)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong, order not found",
		})

		return
	}

	idUser := strconv.Itoa(order.UserID)
	mptoken := os.Getenv("MERCADO_PAGO_ACCESS_TOKEN")
	expirationDate := time.Now().Add(time.Hour * 3)

	response, mercadopagoErr, err := mercadopago.CreatePayment(mercadopago.PaymentRequest{
		ExternalReference: ID,
		Items: []mercadopago.Item{
			{
				ID:        idUser,
				Title:     "Compra de Libros",
				Quantity:  1,
				UnitPrice: order.Total,
			},
		},
		Payer: mercadopago.Payer{
			Name:    order.User.Name,
			Surname: order.User.Last_name,
			Email:   order.User.Email,
		},
		DateOfExpiration: &expirationDate,
		NotificationURL:  "",
	}, mptoken)

	switch {
	case err != nil:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong",
		})
	case mercadopagoErr != nil:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the connection to mercadopago",
		})
		log.Fatal(mercadopagoErr)
	default:
		c.JSON(http.StatusCreated, response)
	}
}
