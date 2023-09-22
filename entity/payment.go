package entity

import "time"

type Payment struct {
	ID             int64            `json:"id"`
	RecipientId    int64            `json:"recipientId,omitempty"`
	PassengerId    int64            `json:"passengerId,omitempty"`
	UserId         int64            `json:"userId,omitempty"`
	Status         string           `json:"status"`
	TotalAmount    int64            `json:"totalAmount"`
	CreatedAt      time.Time        `json:"createdAt"`
	PaymentDetails []PaymentDetails `json:"paymentDetails"`
}

type PaymentDetails struct {
	ID            int64  `json:"id"`
	PaymentId     int64  `json:"paymentId,omitempty"`
	PaymentMethod string `json:"paymentMethod"`
	Amount        int64  `json:"amount"`
}
