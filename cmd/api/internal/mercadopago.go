package internal

import (
	"bookstore-api/internal/business"
	"bookstore-api/internal/platform"
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

	if err != nil || mercadopagoErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp := business.Payment{
		PaymentID: response.ID,
		OrderID:   id,
		UserID:    order.UserID,
		Total:     response.Items[0].UnitPrice,
	}

	db.Create(&resp)
	c.JSON(http.StatusCreated, resp)
}

func (h *Handlers) GetPayment(c *gin.Context) {
	var payment business.Payment

	ID := c.Param("ID")
	id, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	db := platform.DbConnection()
	result := db.Where("Order_id = ?", id).Find(&payment)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	mptoken := os.Getenv("MERCADO_PAGO_ACCESS_TOKEN")
	idPayment := payment.PaymentID

	response, mercadopagoErr, err := mercadopago.GetPayment(idPayment, mptoken)
	switch {
	case err != nil:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong",
		})
	case mercadopagoErr != nil:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the connection to mercadopago",
		})
	default:
		c.JSON(http.StatusOK, response)
	}
}
