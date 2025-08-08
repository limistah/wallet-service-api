package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ReconciliationReport represents a reconciliation report
type ReconciliationReport struct {
	ID                uint                 `json:"id" gorm:"primarykey"`
	CreatedAt         time.Time            `json:"created_at"`
	WalletID          uint                 `json:"wallet_id" gorm:"not null;index"`
	StoredBalance     decimal.Decimal      `json:"stored_balance" gorm:"type:decimal(15,2);not null"`
	CalculatedBalance decimal.Decimal      `json:"calculated_balance" gorm:"type:decimal(15,2);not null"`
	Difference        decimal.Decimal      `json:"difference" gorm:"type:decimal(15,2);not null"`
	Status            ReconciliationStatus `json:"status" gorm:"type:enum('MATCH','MISMATCH','DOUBLE_ENTRY_ERROR');not null"`
	Notes             string               `json:"notes" gorm:"type:text"`

	// Relationships
	Wallet Wallet `json:"wallet,omitempty" gorm:"foreignKey:WalletID"`
}

// ReconciliationStatus represents the status of a reconciliation
type ReconciliationStatus string

const (
	ReconciliationStatusMatch            ReconciliationStatus = "MATCH"
	ReconciliationStatusMismatch         ReconciliationStatus = "MISMATCH"
	ReconciliationStatusDoubleEntryError ReconciliationStatus = "DOUBLE_ENTRY_ERROR"
)

// TableName overrides the table name used by ReconciliationReport
func (ReconciliationReport) TableName() string {
	return "reconciliation_reports"
}

// HasMismatch checks if there's a balance mismatch
func (r *ReconciliationReport) HasMismatch() bool {
	return r.Status == ReconciliationStatusMismatch
}

// HasDoubleEntryError checks if there's a double-entry error
func (r *ReconciliationReport) HasDoubleEntryError() bool {
	return r.Status == ReconciliationStatusDoubleEntryError
}

// HasAnyIssue checks if there's any reconciliation issue
func (r *ReconciliationReport) HasAnyIssue() bool {
	return r.Status != ReconciliationStatusMatch
}

// GetSeverity returns the severity level of the reconciliation issue
func (r *ReconciliationReport) GetSeverity() string {
	switch r.Status {
	case ReconciliationStatusMatch:
		return "INFO"
	case ReconciliationStatusMismatch:
		return "WARNING"
	case ReconciliationStatusDoubleEntryError:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}
