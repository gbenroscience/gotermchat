package cmd

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const ModeCFB = 0
const ModeCBC = 1

// Kryptik When encrypting using this api, make sure your encryption and decryption modes always match!
// So if you encrypted using ModeCFB, alo use ModeCFB when decrypting
type Kryptik struct {
	Key  string
	Mode int //CBC or CFB
}

// NewKryptik creates a new Kryptik pointer. If an invalid mode is specified, returns a nil pointer
func NewKryptik(key string, mode int) (*Kryptik, error) {
	if mode != ModeCFB && mode != ModeCBC {
		return nil, errors.New("invalid AES mode specified... specify mode=0 for CFB and mode=1 for CBC")
	}
	return &Kryptik{
		Key:  key,
		Mode: mode,
	}, nil
}

// DefaultKryptik creates a new Kryptik pointer with CBC mode enabled
func DefaultKryptik(key string, ivText string) *Kryptik {
	return &Kryptik{
		Key:  key,
		Mode: ModeCBC,
	}
}
func (k Kryptik) encryptCFB(message string) (encmess string, err error) {
	if len(k.Key) != 16 && len(k.Key) != 24 && len(k.Key) != 32 {
		return "", errors.New("invalid key size")
	}

	plainText := []byte(message)
	block, err := aes.NewCipher([]byte(k.Key))
	if err != nil {
		return "", err
	}

	// IV needs to be unique, but doesn't have to be secure.
	// It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	// returns to base64 encoded string
	encmess = base64.RawURLEncoding.EncodeToString(cipherText)
	return encmess, nil
}

func (k Kryptik) decryptCFB(encrypted string) (decrypted string, err error) {
	cipherText, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	if len(k.Key) != 16 && len(k.Key) != 24 && len(k.Key) != 32 {
		return "", errors.New("invalid key size")
	}

	block, err := aes.NewCipher([]byte(k.Key))
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("the block size of the ciphertext is too short")
	}

	// IV needs to be unique, but doesn't have to be secure.
	// It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

func (k Kryptik) encryptCBC(message string) (encmess string, err error) {
	// Check if the key size is valid
	if len(k.Key) != 16 && len(k.Key) != 24 && len(k.Key) != 32 {
		return "", errors.New("invalid key size")
	}

	plainText := []byte(message)
	plainTextWithPadding, err := pkcs5Padding(plainText, aes.BlockSize)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(k.Key))
	if err != nil {
		return "", err
	}

	// IV needs to be unique, but doesn't have to be secure.
	// It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainTextWithPadding))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCBCEncrypter(block, iv)
	stream.CryptBlocks(cipherText[aes.BlockSize:], plainTextWithPadding)

	// returns to base64 encoded string
	encmess = base64.RawURLEncoding.EncodeToString(cipherText)
	return encmess, nil
}

func (k Kryptik) decryptCBC(encrypted string) (decrypted string, err error) {
	// Decode the base64 encoded string
	cipherText, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	// Check if the key size is valid
	if len(k.Key) != 16 && len(k.Key) != 24 && len(k.Key) != 32 {
		return "", errors.New("invalid key size")
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher([]byte(k.Key))
	if err != nil {
		return "", err
	}

	// Check if the ciphertext is too short
	if len(cipherText) < aes.BlockSize {
		return "", errors.New("the block size of the ciphertext is too short")
	}

	// Extract the IV from the ciphertext
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// Check if the ciphertext is a multiple of the block size
	if len(cipherText)%aes.BlockSize != 0 {
		return "", errors.New("the ciphertext is not a multiple of the block size")
	}

	// Create a new CBC decrypter
	stream := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the ciphertext
	stream.CryptBlocks(cipherText, cipherText)

	// Remove the padding
	decryptedBytes, err := pkcs5UnPadding(cipherText)
	if err != nil {
		return "", err
	}

	// Convert the decrypted bytes to a string
	decrypted = string(decryptedBytes)

	return decrypted, nil
}

func (k Kryptik) Encrypt(input string) (decrypted string, err error) {
	if k.Mode == ModeCFB {
		return k.encryptCFB(input)
	} else if k.Mode == ModeCBC {
		return k.encryptCBC(input)
	}
	return "", errors.New("invalid encryption mode specified")
}
func (k Kryptik) Decrypt(encrypted string) (decrypted string, err error) {
	if k.Mode == ModeCFB {
		return k.decryptCFB(encrypted)
	} else if k.Mode == ModeCBC {
		return k.decryptCBC(encrypted)
	}
	return "", errors.New("invalid decryption mode specified")
}

func pkcs5Padding(ciphertext []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, errors.New("invalid block size")
	}
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...), nil
}

func pkcs5UnPadding(encrypt []byte) ([]byte, error) {
	blockSize := 16 // Assuming AES block size is 16 bytes
	padding := encrypt[len(encrypt)-1]
	if padding > byte(blockSize) || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	for i := len(encrypt) - int(padding); i < len(encrypt); i++ {
		if encrypt[i] != padding {
			return nil, errors.New("invalid padding")
		}
	}
	return encrypt[:len(encrypt)-int(padding)], nil
}
