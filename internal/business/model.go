package business

import "gorm.io/gorm"

type Product struct {
	ID          int
	Title       string  `gorm:"size:255;not null"`
	Author      string  `gorm:"size:255;not null"`
	Category    string  `gorm:"size:255;not null"`
	Price       float64 `gorm:"not null"`
	Description string
	Language    string
	Cover       string
	Editorial   string
	Year        int
	Pages       int
}

type User struct {
	ID        int
	Email     string `gorm:"size:30;not null;unique" json:"-" binding:"required"`
	Password  string `gorm:"size:255;not null;unique" json:"-" binding:"required"`
	Name      string `gorm:"size:30;not null"`
	Last_name string `gorm:"size:30;not null"`
	Role      string
}

type User_address struct {
	ID     int
	UserID int
	//User        User
	Street      string `gorm:"size:50;not null"`
	Number      int    `gorm:"not null"`
	City        string `gorm:"not null"`
	Postal_code int    `gorm:"not null"`
	Province    string `gorm:"not null"`
}

type Order struct {
	ID            int
	UserID        int
	User          User
	Order_details []Order_details
	Total         float64
	gorm.Model
}

type Order_details struct {
	ID        int
	OrderID   int
	ProductID int
	Product   Product
	Quantity  int
	Total     float64
}

type Payment struct {
	PaymentID string
	OrderID   int
	UserID    int
	Total     float64
}
