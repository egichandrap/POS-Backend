package valueobject

import "fmt"

// Email represents a validated email address
type Email string

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (Email, error) {
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}
	// Basic validation - in production, use a proper email validation library
	if len(email) < 3 || len(email) > 254 {
		return "", fmt.Errorf("invalid email format")
	}
	return Email(email), nil
}

// String returns the email as a string
func (e Email) String() string {
	return string(e)
}
