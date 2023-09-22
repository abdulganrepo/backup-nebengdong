package entity

import "time"

type Passengers struct {
	ID                    int64      `json:"id"`
	UserId                int64      `json:"userId,omitempty"`
	ShareRideId           int64      `json:"shareRideId,omitempty"`
	Status                int16      `json:"status"`
	DestinationCoordinate Coordinate `json:"destinationCoordinate"`
	Distance              float64    `json:"distance"`
	CreatedAt             time.Time  `json:"createdAt"`
	DroppedAt             *time.Time `json:"droppedAt"`
	Payment               []*Payment `json:"payment"`
}
