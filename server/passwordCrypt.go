package main

import (
	"bytes"
	"crypto/rand"
	"log"

	"golang.org/x/crypto/argon2"
)

func hashPass(salt []byte, plainPassword string) []byte {
	hashedPass := argon2.IDKey([]byte(plainPassword), []byte(salt), 1, 64*1024, 4, 32)
	return append(salt, hashedPass...)
}

func checkPass(passHash []byte, plainPassword string) bool {
	salt := passHash[0:8]
	userPassHash := hashPass(salt, plainPassword)
	log.Println(userPassHash)
	return bytes.Equal(userPassHash, passHash)
}

func convertPass(plainPassword string) []byte {
	salt := make([]byte, 8)
	rand.Read(salt)
	return hashPass(salt, plainPassword)
}
