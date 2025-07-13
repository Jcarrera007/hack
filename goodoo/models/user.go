package models

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	Login     string `gorm:"unique;not null" json:"login"`
	Name      string `gorm:"" json:"name"`
	Email     string `gorm:"unique" json:"email"`
	Password  string `gorm:"" json:"-"`
	Active    bool   `gorm:"default:true" json:"active"`
	PartnerID *uint  `gorm:"column:partner_id" json:"partner_id,omitempty"`
	Share     bool   `gorm:"default:false" json:"share"`
}

func (User) TableName() string {
	return "res_users"
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// SetPasswordOdooStyle creates an Odoo-compatible PBKDF2-SHA512 password hash
func (u *User) SetPasswordOdooStyle(password string) error {
	const rounds = 600000 // Odoo default minimum rounds
	
	// Generate random salt (16 bytes)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	
	// Generate hash
	hash := pbkdf2.Key([]byte(password), salt, rounds, 32, sha512.New)
	
	// Format as Odoo-style hash: $pbkdf2-sha512$rounds$salt$hash
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	hashB64 := base64.StdEncoding.EncodeToString(hash)
	
	u.Password = fmt.Sprintf("$pbkdf2-sha512$%d$%s$%s", rounds, saltB64, hashB64)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	if u.Password == "" {
		return false
	}

	// Handle Odoo PBKDF2-SHA512 format first
	if strings.HasPrefix(u.Password, "$pbkdf2-sha512$") {
		return u.verifyPBKDF2Password(password)
	}

	// Handle bcrypt format (our own created users)
	if strings.HasPrefix(u.Password, "$2") {
		err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
		return err == nil
	}

	// Handle plaintext (legacy, not recommended)
	return u.Password == password
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// verifyPBKDF2Password verifies Odoo-style PBKDF2-SHA512 passwords
func (u *User) verifyPBKDF2Password(password string) bool {
	// Odoo format: $pbkdf2-sha512$rounds$salt$hash
	parts := strings.Split(u.Password, "$")
	if len(parts) != 5 || parts[0] != "" || parts[1] != "pbkdf2-sha512" {
		return false
	}

	rounds, err := strconv.Atoi(parts[2])
	if err != nil {
		return false
	}

	// Helper function to decode passlib's adapted base64 format
	decodePasslibBase64 := func(s string) ([]byte, error) {
		// Convert passlib's adapted base64 format to standard base64
		// In passlib MCF format: dots (.) are used instead of plus (+)
		s = strings.ReplaceAll(s, ".", "+")
		
		// Add padding if needed
		missing := len(s) % 4
		if missing != 0 {
			s += strings.Repeat("=", 4-missing)
		}
		
		// Try standard base64 first
		if data, err := base64.StdEncoding.DecodeString(s); err == nil {
			return data, nil
		}
		
		// Try URL-safe base64 as fallback
		if data, err := base64.URLEncoding.DecodeString(s); err == nil {
			return data, nil
		}
		
		return nil, fmt.Errorf("invalid base64 data")
	}

	salt, err := decodePasslibBase64(parts[3])
	if err != nil {
		return false
	}

	expectedHash, err := decodePasslibBase64(parts[4])
	if err != nil {
		return false
	}

	// Generate hash with same parameters
	actualHash := pbkdf2.Key([]byte(password), salt, rounds, len(expectedHash), sha512.New)

	// Constant time comparison
	if len(actualHash) != len(expectedHash) {
		return false
	}

	var result byte
	for i := 0; i < len(actualHash); i++ {
		result |= actualHash[i] ^ expectedHash[i]
	}
	
	return result == 0
}

func FindUserByLogin(db *gorm.DB, login string) (*User, error) {
	var user User
	err := db.Where("login = ? AND active = ?", login, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(db *gorm.DB, login, name, email, password string) (*User, error) {
	user := &User{
		Login: login,
		Name:  name,
		Email: email,
		Active: true,
	}
	
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	
	err := db.Create(user).Error
	if err != nil {
		return nil, err
	}
	
	return user, nil
}