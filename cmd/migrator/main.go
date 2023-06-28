package main

import (
	"bookstore-api/internal/business"
	"bookstore-api/internal/platform"
)

func main() {
	db := platform.DbConnection()
	db.Migrator().CreateTable(&business.User{}, &business.User_details{}, &business.Order{}, &business.Order_detail{}, &business.Payment{})

}
