package business

import "gorm.io/gorm"

type Product struct {
	ID          int
	Title       string
	Author      string
	Category    string
	Price       float64
	Description string
}

type User struct {
	ID        int
	Email     string `json:"-"`
	Password  string `json:"-"`
	Name      string
	Last_name string
}

type User_details struct {
	ID          int
	UserID      int
	User        User
	Street      string
	Number      int
	City        string
	Postal_code int
	Province    string
}

type Order struct {
	ID     int
	UserID int
	User   User
	Total  int
}

type Order_detail struct {
	ID        int
	OrderID   int
	Order     Order
	ProductID int
	Product   Product
	Quantity  int
}

type Payment struct {
	gorm.Model
	OrderID int
	Order   Order
}
