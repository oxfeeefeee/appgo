package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"io"
	"math/big"
)

const (
	numberchars = "1234567890"
)

func RandNumStr(l int) string {
	var max big.Int
	max.SetInt64(int64(len(numberchars)))
	ret := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		index, _ := rand.Int(rand.Reader, &max)
		ret = append(ret, numberchars[index.Int64()])
	}
	return string(ret)
}

func RandBytes(c int) ([]byte, error) {
	b := make([]byte, c)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}

func SaltedHash(salt []byte, password string) [64]byte {
	pw := append(salt, []byte(password)...)
	return sha512.Sum512(pw)
}

func Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := padding(data)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return appendHmac(ciphertext, key), nil
}

func Decrypt(data, key []byte) ([]byte, error) {
	data, err := removeHmac(data, key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	// CBC mode always works in whole blocks.
	if len(data)%aes.BlockSize != 0 {
		errors.New("data is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)
	return unPadding(plaintext), nil
}

func padding(data []byte) []byte {
	blockSize := aes.BlockSize
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func unPadding(data []byte) []byte {
	l := len(data)
	unpadding := int(data[l-1])
	return data[:(l - unpadding)]
}

func appendHmac(data, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	hash := mac.Sum(nil)
	return append(data, hash...)
}

func removeHmac(data, key []byte) ([]byte, error) {
	if len(data) < sha256.Size {
		return nil, errors.New("Invalid length")
	}
	p := len(data) - sha256.Size
	mmac := data[p:]
	mac := hmac.New(sha256.New, key)
	mac.Write(data[:p])
	exp := mac.Sum(nil)
	if hmac.Equal(mmac, exp) {
		return data[:p], nil
	} else {
		return nil, errors.New("MAC doesn't match")
	}
}
