package internal

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func GenerateToken(user_id int) (string, error) {
	token_lifespan, err := strconv.Atoi(os.Getenv("TOKEN_HOUR_LIFESPAN"))

	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = user_id
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(token_lifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return (token.SignedString([]byte(os.Getenv("TOKEN_KEY"))))
}

func TokenValid(c *gin.Context) error {
	tokenString := ExtractToken(c)
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("TOKEN_KEY")), nil
	})

	if err != nil {
		return err
	}

	user_id, ok := (claims["user_id"]).(float64)

	if !ok {
		return errors.New("user Id not exists")
	}

	c.Set("user_id", user_id)
	return nil
}

func ExtractToken(c *gin.Context) string {

	bearerToken := c.Request.Header.Get("Authorization")
	bToken := strings.Split(bearerToken, " ")
	if len(bToken) == 2 {
		return bToken[1]
	}
	return ""
}
