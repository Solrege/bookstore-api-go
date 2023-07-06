package main

import (
	"bookstore-api/internal/business"
	"bookstore-api/internal/platform"
)

func main() {
	db := platform.DbConnection()
	db.Migrator().AddColumn(&business.User{}, "Role")
}
