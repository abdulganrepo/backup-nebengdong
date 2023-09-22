package model

type VehicleRegistration struct {
	VehicleModel        string `json:"vehicleModel" binding:"required"`
	VehicleManufature   string `json:"vehicleManufature" binding:"required"`
	VehicleLicensePlate string `json:"vehicleLicensePlate" binding:"required,max=9"`
}
