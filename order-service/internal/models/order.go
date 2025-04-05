package models

var (
	AllowPaymentMethod = map[string]bool{
		"stripe": true,
	}

	AllowPaymentStatus = map[string]bool{
		"success":    true,
		"cancel":     true,
		"processing": true,
		"refunded":   true,
	}
)

type Order struct {
	BaseModel
	UserID        string        `gorm:"type:varchar(36);not null;index"`
	PaymentMethod string        `gorm:"type:varchar(20);not null;uniqueIndex:idx_unique_transaction_payment"`
	TransactionID string        `gorm:"type:varchar(255);not null;uniqueIndex:idx_unique_transaction_payment"`
	AmountPaid    float64       `gorm:"type:float;not null"`
	PaymentStatus string        `gorm:"type:varchar(20);not null;default:'processing';index"`
	ChargeID      string        `gorm:"type:varchar(255);not null"`
	OrderDetails  []OrderDetail `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
}
