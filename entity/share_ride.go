package entity

import "time"

type ShareRide struct {
	ID           int64          `json:"id"`
	DriverId     int64          `json:"driverId,omitempty"`
	IsFull       bool           `json:"isFull"`
	DriverStatus int16          `json:"driverStatus"`
	CreatedAt    time.Time      `json:"createdAt"`
	FinishedAt   *time.Time     `json:"finishedAt"`
	Passengers   []*Passengers  `json:"passengers"`
	Users        *UserInVehicle `json:"users"`
	Driver       *UserInVehicle `json:"driver"`
}
