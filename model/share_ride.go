package model

type ShareRideId struct {
	ID int64 `json:"id" binding:"min=1,number"`
}

type FinishFindPassenger struct {
	ShareRideId
}

type FindDriver struct {
	DestinationCoordinate Coordinate `json:"destinationCoordinate" binding:"required"`
	Distance              float64    `json:"distance" binding:"required"`
	CostPerKM             int64      `json:"costPerKm" binding:"required"`
}
