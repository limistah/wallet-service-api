package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// System account constants
const (
	SystemAccountEmail = "system@wallet.internal"
	SystemAccountName  = "System Account"
)

type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Name      string         `json:"name" gorm:"type:varchar(255);not null" validate:"required,min=2,max=100"`
	Email     string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null" validate:"required,email"`
	Password  string         `json:"-" gorm:"type:varchar(255);not null" validate:"required,min=6"` // "-" excludes from JSON serialization
	Age       int            `json:"age" validate:"omitempty,gte=0,lte=150"`
	IsSystem  bool           `json:"is_system" gorm:"default:false;index"` // For system accounts

	// Relationships
	Wallets []Wallet `json:"wallets,omitempty" gorm:"foreignKey:UserID"`
}

// TableName overrides the table name used by User to `users`
func (User) TableName() string {
	return "users"
}

// HashPassword hashes the user's password using bcrypt
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies the password against the hashed password
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// IsSystemAccount checks if this is a system account
func (u *User) IsSystemAccount() bool {
	return u.IsSystem
}

// CreateSystemUser creates a system user instance
func CreateSystemUser() *User {
	return &User{
		Name:     SystemAccountName,
		Email:    SystemAccountEmail,
		Password: "system-account-password", // This will be hashed
		IsSystem: true,
	}
}
