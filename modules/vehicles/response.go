package vehicles

import (
	"time"

	"github.com/Difaal21/nebeng-dong/entity"
)

type vehicleResponses struct {
	ID           int64                 `json:"id"`
	Users        *entity.UserInVehicle `json:"user"`
	Type         string                `json:"type"`
	Model        string                `json:"model"`
	LicensePlate string                `json:"licensePlate"`
	Manufacture  string                `json:"manufacture"`
	InUse        bool                  `json:"inUse"`
	Capacity     int                   `json:"capacity"`
	CreatedAt    time.Time             `json:"createdAt"`
}
