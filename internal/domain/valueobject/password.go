package valueobject

import "fmt"

// Password represents a hashed password
type Password string

// NewPassword creates a new Password value object
func NewPassword(password string) (Password, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	if len(password) < 6 {
		return "", fmt.Errorf("password must be at least 6 characters")
	}
	return Password(password), nil
}

// String returns the password hash as a string
func (p Password) String() string {
	return string(p)
}

// IsEmpty checks if the password is empty
func (p Password) IsEmpty() bool {
	return p == ""
}
