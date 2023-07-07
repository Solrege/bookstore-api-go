package internal

import (
	"bookstore-api/internal/business"
	"bookstore-api/internal/platform"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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

type createOrderRequest struct {
	Order_details []orderDetailRequest
}

type orderDetailRequest struct {
	ProductID int
	Quantity  int
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

	mUser := business.User{
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

	c.JSON(http.StatusCreated, user.Email)
}

func verifyPassword(hashPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
}

func (h *Handlers) LoginHandler(c *gin.Context) {
	var input loginRequest
	var user business.User

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

	resp, err := GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Something went wrong with the token: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) GetBookByIDHandler(c *gin.Context) {
	ID := c.Param("ID")

	id, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the ID",
		})

		return
	}

	var book business.Product

	db := platform.DbConnection()

	result := db.Where("ID = ?", id).Find(&book)
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
	sortDirection := c.Query("dir")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	book := []business.Product{}

	db := platform.DbConnection()
	result := db.Where("category = ?", category).Find(&book)

	//sort by author o title
	order := "ASC"
	if sortDirection == "1" {
		order = "DESC"
	}

	switch sortBy {
	case "title":
		result = result.Order("title " + order)
	default:
		result = result.Order("author " + order)
	}

	//pagination
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong",
		})

		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong",
		})

		return
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

func (h *Handlers) GetBooksByAuthorHandler(c *gin.Context) {
	author := c.Param("author")
	book := []business.Product{}

	db := platform.DbConnection()

	result := db.Where("author = ?", author).Order("title ASC").Find(&book)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "There is not book for the Author",
		})

		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *Handlers) SearchBookHandler(c *gin.Context) {
	originalQuery := c.Query("query")

	//quitar espacios
	query := strings.TrimSpace(originalQuery)

	// longitud
	q := len(query)
	if q == 0 || q > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid query",
		})
	}

	// evitar caract especiales
	reg := regexp.MustCompile("[^a-zA-Z0-9 ]+")

	cleanQuery := reg.ReplaceAllString(query, "")

	var book []business.Product

	db := platform.DbConnection()

	result := db.Where("title LIKE ? OR author LIKE ?", "%"+cleanQuery+"%", "%"+cleanQuery+"%").Find(&book)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Book or Author not found",
		})

		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *Handlers) AddNewBookHandler(c *gin.Context) {
	var book business.Product

	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	db := platform.DbConnection()
	result := db.Create(&book)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Book not created",
		})
		return
	}

	c.JSON(http.StatusCreated, book)
}

func (h *Handlers) DeleteBookHandler(c *gin.Context) {
	ID := c.Param("ID")

	id, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the ID",
		})

		return
	}

	var book business.Product

	db := platform.DbConnection()
	delete := db.Where("ID = ?", id).Delete(&book)

	if delete.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Book not deleted",
		})
		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *Handlers) UpdateBookHandler(c *gin.Context) {
	ID := c.Param("ID")

	id, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Something went wrong with the ID",
		})

		return
	}

	var book business.Product

	db := platform.DbConnection()

	// primero se busca el libro
	result := db.Find(&book, id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Book not found",
		})

		return
	}

	// se bindea
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	// se actualiza
	update := db.Model(&book).Updates(book)

	if update.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Book not updated",
		})

		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *Handlers) GetAddressHandler(c *gin.Context) {
	var userAddress business.User_address

	user_id, _ := c.Get("user_id")
	userId := user_id.(float64)
	userAddress.UserID = int(userId)

	db := platform.DbConnection()

	result := db.Find(&userAddress)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User address not found",
		})

		return
	}

	c.JSON(http.StatusOK, userAddress)
}

func (h *Handlers) AddAddressHandler(c *gin.Context) {
	var userAddress business.User_address

	if err := c.ShouldBindJSON(&userAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	user_id, _ := c.Get("user_id")
	userId := user_id.(float64)
	userAddress.UserID = int(userId)

	db := platform.DbConnection()
	result := db.Create(&userAddress)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User address not created",
		})

		return
	}

	c.JSON(http.StatusCreated, userAddress)
}

func (h *Handlers) UpdateAddressHandler(c *gin.Context) {
	var userAddress business.User_address

	user_id, _ := c.Get("user_id")
	userId := user_id.(float64)
	userAddress.UserID = int(userId)

	db := platform.DbConnection()

	result := db.Find(&userAddress)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Address not found",
		})

		return
	}

	if err := c.ShouldBindJSON(&userAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	update := db.Model(&userAddress).Updates(&userAddress)

	if update.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Book not updated",
		})

		return
	}

	c.JSON(http.StatusOK, userAddress)
}

func (h *Handlers) CreateOrderHandler(c *gin.Context) {
	var product business.Product
	var preOrder createOrderRequest

	if err := c.ShouldBindJSON(&preOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	db := platform.DbConnection()

	user_id, _ := c.Get("user_id")
	userId := user_id.(float64)

	order := business.Order{
		UserID: int(userId),
		Total:  0,
	}

	items := make([]business.Order_details, 0)
	for _, v := range preOrder.Order_details {

		product.ID = v.ProductID
		result := db.Where("ID = ?", v.ProductID).Find(&product)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Something went wrong with the search",
			})

			return
		}

		precio := product.Price
		subtotal := precio * float64(v.Quantity)

		item := business.Order_details{
			ProductID: v.ProductID,
			Quantity:  v.Quantity,
			Total:     subtotal,
		}

		order.Total += subtotal

		items = append(items, item)
	}

	order.Order_details = items

	db.Create(&order)

	c.JSON(http.StatusCreated, order)
}
