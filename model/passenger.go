package model

type UpdatePassengerStatus struct {
	ID          int64 `json:"id" binding:"required,min=1"`
	Code        int8  `json:"code" binding:"required,number"`
	ShareRideID int64 `json:"shareRideID" binding:"required,min=1"`
}

type TopUpCoinBalance struct {
	ID   int64 `json:"id" binding:"required,min=1"`
	Coin int64 `json:"coin" binding:"required,min=1"`
}
