package users

import (
	"fmt"

	"github.com/Difaal21/nebeng-dong/entity"
)

func VehicleNullHandler(vehicle *entity.Vehicle) (err error) {
	if vehicle.LicensePlate == "" || vehicle.Manufacture == "" || vehicle.Model == "" {
		return fmt.Errorf("invalid payload of vehicle")
	}

	return
}
