package entity

import "time"

type ShareRide struct {
	ID           int64              `json:"id"`
	DriverId     int64              `json:"driverId,omitempty"`
	IsFull       bool               `json:"isFull"`
	DriverStatus int16              `json:"driverStatus"`
	CreatedAt    time.Time          `json:"createdAt"`
	FinishedAt   *time.Time         `json:"finishedAt"`
	Passengers   []*Passengers      `json:"passengers"`
	Driver       *DriverInShareRide `json:"driver"`
}

type DriverInShareRide struct {
	ID          int64                     `json:"id"`
	Name        string                    `json:"name"`
	Email       string                    `json:"email"`
	PhoneNumber string                    `json:"phoneNumber"`
	Vehicle     *DriverVehicleInShareRide `json:"vehicle,omitempty"`
}

type DriverVehicleInShareRide struct {
	ID           int64  `json:"id"`
	Model        string `json:"model"`
	LicensePlate string `json:"licensePlate"`
	Manufacture  string `json:"manufacture"`
	InUse        bool   `json:"inUse"`
}
