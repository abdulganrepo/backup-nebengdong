package model

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Identifier struct{}

func GetRequester(ctx context.Context) (user *UserBearer, err error) {
	userContext := ctx.Value(&Identifier{})
	user, ok := userContext.(*UserBearer)
	if !ok {
		err = fmt.Errorf("invalid context key")
		return
	}

	return
}

type UserRegistration struct {
	Name                string `json:"name" binding:"required"`
	Email               string `json:"email" binding:"required,email"`
	PhoneNumber         string `json:"phoneNumber" binding:"required"`
	Password            string `json:"password" binding:"required"`
	IsDriver            *bool  `json:"isDriver" binding:"required"`
	VehicleModel        string `json:"vehicleModel" binding:"omitempty"`
	VehicleManufature   string `json:"vehicleManufature" binding:"omitempty"`
	VehicleLicensePlate string `json:"vehicleLicensePlate" binding:"omitempty,max=9"`
}

type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserBearer struct {
	jwt.RegisteredClaims
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsDriver bool   `json:"isDriver"`
}

type Coordinate struct {
	Latitude  float64 `json:"lat" binding:"required,latitude"`
	Longitude float64 `json:"long" binding:"required,longitude"`
}

type ChangePhoneNumber struct {
	New string `json:"new" binding:"required"`
}

type ChangePassword struct {
	New string `json:"new" binding:"required"`
	Old string `json:"old" binding:"required"`
}
