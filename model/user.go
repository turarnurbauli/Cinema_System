package model

// User представляет пользователя системы кинотеатра.
// Role: "customer", "cashier", "admin".
type User struct {
	ID           int    `json:"id" bson:"id"`
	Email        string `json:"email" bson:"email"`
	PasswordHash string `json:"-" bson:"password_hash"`
	Name         string `json:"name" bson:"name"`
	Role         string `json:"role" bson:"role"`
}

