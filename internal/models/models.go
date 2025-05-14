package models

import "time"

// Category represents a pet category.
type Category struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// Tag represents a pet tag.
type Tag struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// Pet represents a pet in the store.
type Pet struct {
	ID        int64     `db:"id" json:"id"`
	Category  *Category `json:"category"`
	Name      string    `db:"name" json:"name"`
	PhotoURLs []string  `json:"photoUrls"`
	Tags      []Tag     `json:"tags"`
	Status    string    `db:"status" json:"status"`
}

// Order represents a purchase order for a pet.
type Order struct {
	ID       int64     `db:"id" json:"id"`
	PetID    int64     `db:"pet_id" json:"petId"`
	Quantity int       `db:"quantity" json:"quantity"`
	ShipDate time.Time `db:"ship_date" json:"shipDate"`
	Status   string    `db:"status" json:"status"`
	Complete bool      `db:"complete" json:"complete"`
}

// User represents a user in the system.
type User struct {
	ID         int64  `db:"id" json:"id"`
	Username   string `db:"username" json:"username"`
	FirstName  string `db:"first_name" json:"firstName"`
	LastName   string `db:"last_name" json:"lastName"`
	Email      string `db:"email" json:"email"`
	Password   string `db:"password" json:"password"`
	Phone      string `db:"phone" json:"phone"`
	UserStatus int    `db:"user_status" json:"userStatus"`
}
