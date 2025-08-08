package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TransactionType represents the type of transaction
type TransactionType struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name" gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string    `json:"description" gorm:"type:text"`

	// Relationships
	Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:TransactionTypeID"`
}

// Transaction type constants
const (
	TransactionTypeCredit = "CREDIT"
	TransactionTypeDebit  = "DEBIT"
)

// TableName overrides the table name used by TransactionType
func (TransactionType) TableName() string {
	return "transaction_types"
}

// TransactionPurpose represents the type of transaction
type TransactionPurpose string

// Transaction represents a wallet transaction
type Transaction struct {
	ID                   uint              `json:"id" gorm:"primarykey"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	DeletedAt            gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index"`
	Reference            string            `json:"reference" gorm:"type:varchar(255);uniqueIndex;not null"`
	WalletID             uint              `json:"wallet_id" gorm:"not null;index"`
	TransactionTypeID    uint              `json:"transaction_type_id" gorm:"not null;index"`
	Amount               decimal.Decimal   `json:"amount" gorm:"type:decimal(15,2);not null;check:amount > 0"`
	BalanceBefore        decimal.Decimal   `json:"balance_before" gorm:"type:decimal(15,2);not null"`
	BalanceAfter         decimal.Decimal   `json:"balance_after" gorm:"type:decimal(15,2);not null"`
	Description          string            `json:"description" gorm:"type:text"`
	Metadata             string            `json:"metadata" gorm:"type:json"`
	Status               TransactionStatus `json:"status" gorm:"type:enum('PENDING','COMPLETED','FAILED','CANCELLED');not null;default:'PENDING'"`
	RelatedTransactionID *uint             `json:"related_transaction_id,omitempty" gorm:"index"`

	// Relationships
	Wallet             Wallet             `json:"wallet,omitempty" gorm:"foreignKey:WalletID"`
	TransactionType    TransactionType    `json:"transaction_type,omitempty" gorm:"foreignKey:TransactionTypeID"`
	TransactionPurpose TransactionPurpose `json:"transaction_purpose,omitempty" gorm:"type:enum('WITHDRAWAL','WALLET_TOP_UP','TRANSFER');not null;"`
	RelatedTransaction *Transaction       `json:"related_transaction,omitempty" gorm:"foreignKey:RelatedTransactionID"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

// TableName overrides the table name used by Transaction
func (Transaction) TableName() string {
	return "transactions"
}

// IsCompleted checks if the transaction is completed
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}
