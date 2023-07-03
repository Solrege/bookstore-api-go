package business

import "gorm.io/gorm"

type Products struct {
	ID       int
	Title    string
	Author   string
	Category string
	Price    float64
}

type Product_details struct {
	ID          int
	ProductID   int
	Product     Products
	Description string
	Idioma      string
	Tapa        string
	Editorial   string
	Año         int
	Páginas     int
}

type Users struct {
	ID        int
	Email     string `json:"-"`
	Password  string `json:"-"`
	Name      string
	Last_name string
}

type User_details struct {
	ID          int
	UserID      int
	User        Users
	Street      string
	Number      int
	City        string
	Postal_code int
	Province    string
}

type Orders struct {
	ID     int
	UserID int
	User   Users
	Total  int
}

type Order_details struct {
	ID        int
	OrderID   int
	Order     Orders
	ProductID int
	Product   Products
	Quantity  int
}

type Payments struct {
	gorm.Model
	OrderID int
	Order   Orders
}
