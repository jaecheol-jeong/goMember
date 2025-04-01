package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Member struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *Member) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	m.Password = string(hashedPassword)
	return nil
}

func (m *Member) ComparePasswords(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(password))
}
