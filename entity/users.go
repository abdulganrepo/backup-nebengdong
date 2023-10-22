package entity

import (
	"time"
)

type Users struct {
	ID              int64             `json:"id"`
	Name            string            `json:"name"`
	Email           string            `json:"email"`
	PhoneNumber     string            `json:"phoneNumber"`
	Password        *string           `json:"password,omitempty"`
	Coin            int64             `json:"coin"`
	Coordinate      *Coordinate       `json:"coordinate"`
	IsEmailVerified bool              `json:"isEmailVerified"`
	EmailVerifiedAt *time.Time        `json:"emailVerifiedAt"`
	IsDriver        bool              `json:"isDriver"`
	Vehicles        []*VehiclesInUser `json:"vehicles,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       *time.Time        `json:"updatedAt"`
}

type VehiclesInUser struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	Model        string    `json:"model"`
	LicensePlate string    `json:"licensePlate"`
	Manufacture  string    `json:"manufacture"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
