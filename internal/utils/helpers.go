package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidateEmail validates email format using regex
func ValidateEmail(email string) bool {
	// Check length first (RFC 5321 limits email to 320 characters total)
	if len(email) > 320 {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePassword validates password strength
func ValidatePassword(password string) bool {
	// Minimum 6 characters, not empty, not only spaces
	if len(strings.TrimSpace(password)) < 6 {
		return false
	}
	return true
}

// ValidateAmount validates that amount is positive
func ValidateAmount(amount float64) bool {
	return amount > 0
}

// SanitizeString trims whitespace from string
func SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

// NormalizeEmail converts email to lowercase and trims spaces
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// GenerateReference generates a unique reference string
func GenerateReference() string {
	timestamp := time.Now().Unix()
	randomPart := generateRandomString(6)
	return fmt.Sprintf("REF-%d-%s", timestamp, randomPart)
}

// GenerateTransactionReference generates a transaction reference
func GenerateTransactionReference() string {
	timestamp := time.Now().Unix()
	randomPart := generateRandomString(8)
	return fmt.Sprintf("TXN-%d-%s", timestamp, randomPart)
}

// FormatCurrency formats amount with currency symbol
func FormatCurrency(amount float64, currency string) string {
	formattedAmount := formatAmountWithCommas(amount)

	switch currency {
	case "USD":
		if amount < 0 {
			return fmt.Sprintf("-$%s", formatAmountWithCommas(-amount))
		}
		return fmt.Sprintf("$%s", formattedAmount)
	case "EUR":
		if amount < 0 {
			return fmt.Sprintf("-€%s", formatAmountWithCommas(-amount))
		}
		return fmt.Sprintf("€%s", formattedAmount)
	case "GBP":
		if amount < 0 {
			return fmt.Sprintf("-£%s", formatAmountWithCommas(-amount))
		}
		return fmt.Sprintf("£%s", formattedAmount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

// IsValidCurrency checks if currency code is valid
func IsValidCurrency(currency string) bool {
	validCurrencies := map[string]bool{
		"USD": true,
		"EUR": true,
		"GBP": true,
		"NGN": true,
		"CAD": true,
		"AUD": true,
		"JPY": true,
		"CHF": true,
	}

	return len(currency) == 3 && validCurrencies[currency]
}

// Helper function to generate random string
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

// Helper function to format amount with commas
func formatAmountWithCommas(amount float64) string {
	// Convert to string with 2 decimal places
	amountStr := fmt.Sprintf("%.2f", amount)

	// Split into integer and decimal parts
	parts := strings.Split(amountStr, ".")
	integerPart := parts[0]
	decimalPart := parts[1]

	// Add commas to integer part
	if len(integerPart) > 3 {
		var result []string
		for i, digit := range reverse(integerPart) {
			if i > 0 && i%3 == 0 {
				result = append(result, ",")
			}
			result = append(result, string(digit))
		}
		integerPart = reverse(strings.Join(result, ""))
	}

	return integerPart + "." + decimalPart
}

// Helper function to reverse string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ParseAmount converts string to float64
func ParseAmount(amountStr string) (float64, error) {
	return strconv.ParseFloat(amountStr, 64)
}

// GenerateUniqueID generates a unique identifier
func GenerateUniqueID() string {
	timestamp := time.Now().UnixNano()
	randomPart := generateRandomString(4)
	return fmt.Sprintf("%d%s", timestamp, randomPart)
}
