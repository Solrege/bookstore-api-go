package internal

import (
	"bookstore-api/internal/business"
	"bookstore-api/internal/platform"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
}

func (h *Handlers) Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "holis",
	})

}

type registerRequest struct {
	Email             string
	Password          string
	Pass_confirmation string
	Name              string
	Last_name         string
}

type loginRequest struct {
	Email    string
	Password string
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (h *Handlers) RegisterHandler(c *gin.Context) {
	var user registerRequest

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	db := platform.DbConnection()

	if user.Password != user.Pass_confirmation {
		return
	}

	var err error

	user.Password, err = HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the hash",
		})
	}

	mUser := business.Users{
		Email:     user.Email,
		Password:  user.Password,
		Name:      user.Name,
		Last_name: user.Last_name,
	}

	result := db.Create(&mUser)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func verifyPassword(hashPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
}

func (h *Handlers) LoginHandler(c *gin.Context) {
	var input loginRequest
	var user business.Users

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	db := platform.DbConnection()

	result := db.Model(user).Where("email = ?", input.Email).Take(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the Email",
		})
		return
	}
	fmt.Printf("%v", user)

	err := verifyPassword(user.Password, input.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the password",
		})
		return
	}

	fmt.Println(GenerateToken(user.ID))

	c.JSON(http.StatusOK, input)
}

func (h *Handlers) GetBookByIDHandler(c *gin.Context) {
	ID := c.Param("ID")

	var book []business.Products

	db := platform.DbConnection()

	result := db.Where("ID = ?", ID).Find(&book)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the ID",
		})
		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *Handlers) GetBooksByCategoryHandler(c *gin.Context) {
	category := c.Param("category")
	sortBy := c.Query("sort")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	var book []business.Products

	db := platform.DbConnection()
	result := db.Where("category = ?", category)

	//sort by author o title
	if sortBy == "author" {
		result = result.Order("author ASC")
	} else if sortBy == "title" {
		result = result.Order("title ASC")
	}

	//pagination
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		panic(err)
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		panic(err)
	}

	offset := (page - 1) * limit
	result = result.Offset(offset).Limit(limit)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, book)
}
