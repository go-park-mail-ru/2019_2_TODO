package utils

import (
	"bytes"
	"crypto/rand"
	"log"

	"golang.org/x/crypto/argon2"
)

// HashPass - return slice byte of out hashed pass with argon2
func HashPass(salt []byte, plainPassword string) []byte {
	hashedPass := argon2.IDKey([]byte(plainPassword), []byte(salt), 1, 64*1024, 4, 32)
	return append(salt, hashedPass...)
}

// CheckPass - check equals of 2 hashes
func CheckPass(passHash []byte, plainPassword string) bool {
	salt := passHash[0:8]
	userPassHash := HashPass(salt, plainPassword)
	log.Println(userPassHash)
	return bytes.Equal(userPassHash, passHash)
}

// ConvertPass - conver pass into argon2 hashed pass with random salt
func ConvertPass(plainPassword string) []byte {
	salt := make([]byte, 8)
	rand.Read(salt)
	return HashPass(salt, plainPassword)
}
