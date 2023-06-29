package main

import (
	"bookstore-api/internal/business"
	"bookstore-api/internal/platform"
)

func main() {
	db := platform.DbConnection()
	db.Migrator().CreateTable(&business.Products{}, &business.Users{}, &business.User_details{}, &business.Orders{}, &business.Order_details{}, &business.Payments{})

}
