package entity

import "time"

type Vehicle struct {
	ID           int64          `json:"id"`
	Users        *UserInVehicle `json:"user"`
	UserId       int64          `json:"userId"`
	Type         string         `json:"type"`
	Model        string         `json:"model"`
	LicensePlate string         `json:"licensePlate"`
	Manufacture  string         `json:"manufacture"`
	InUse        bool           `json:"inUse"`
	Capacity     int            `json:"capacity"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type UserInVehicle struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}
