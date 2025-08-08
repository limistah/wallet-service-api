package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Wallet represents a user's wallet
type Wallet struct {
	ID        uint            `json:"id" gorm:"primarykey"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index"`
	UserID    uint            `json:"user_id" gorm:"not null;index"`
	Balance   decimal.Decimal `json:"balance" gorm:"type:decimal(15,2);not null;default:0.00;check:balance >= 0"`
	Currency  string          `json:"currency" gorm:"type:varchar(3);not null;default:'USD'"`
	Status    WalletStatus    `json:"status" gorm:"type:enum('ACTIVE','SUSPENDED','CLOSED');not null;default:'ACTIVE'"`
	Version   uint            `json:"version" gorm:"not null;default:0"` // For optimistic locking

	// Relationships
	User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:WalletID"`
}

// WalletStatus represents the status of a wallet
type WalletStatus string

const (
	WalletStatusActive    WalletStatus = "ACTIVE"
	WalletStatusSuspended WalletStatus = "SUSPENDED"
	WalletStatusClosed    WalletStatus = "CLOSED"
)

// TableName overrides the table name used by Wallet
func (Wallet) TableName() string {
	return "wallets"
}

// IsActive checks if the wallet is active
func (w *Wallet) IsActive() bool {
	return w.Status == WalletStatusActive
}

// CanDebit checks if the wallet can be debited by the specified amount
func (w *Wallet) CanDebit(amount decimal.Decimal) bool {
	return w.IsActive() && w.Balance.GreaterThanOrEqual(amount)
}
