package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)


func generateECCKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}


func encodePrivateKey(privateKey *ecdsa.PrivateKey) (string, error) {
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(keyBytes), nil
}


func encodePublicKey(publicKey *ecdsa.PublicKey) (string, error) {
	keyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(keyBytes), nil
}


func decodePrivateKey(encodedKey string) (*ecdsa.PrivateKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(keyBytes)
}


func decodePublicKey(encodedKey string) (*ecdsa.PublicKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, err
	}
	pubKey, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}
	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}
	return ecdsaPubKey, nil
}


func encryptWithPublicKey(data []byte, publicKey *ecdsa.PublicKey) (string, error) {
	
	ecdhPublicKey, err := publicKey.ECDH()
	if err != nil {
		return "", err
	}

	
	ephemeralPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}

	
	sharedSecret, err := ephemeralPriv.ECDH(ecdhPublicKey)
	if err != nil {
		return "", err
	}

	
	hashedSecret := sha256.Sum256(sharedSecret)

	
	block, err := aes.NewCipher(hashedSecret[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	
	ephemeralPubBytes := ephemeralPriv.PublicKey().Bytes()

	
	result := append(ephemeralPubBytes, nonce...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}


func decryptWithPrivateKey(encryptedData string, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	dataBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	
	if len(dataBytes) < 65+12+16 { 
		return nil, fmt.Errorf("invalid encrypted data size")
	}

	ephemeralPubBytes := dataBytes[:65]
	
	
	ephemeralECDHPub, err := ecdh.P256().NewPublicKey(ephemeralPubBytes)
	if err != nil {
		return nil, err
	}

	
	ecdhPrivateKey, err := privateKey.ECDH()
	if err != nil {
		return nil, err
	}

	
	sharedSecret, err := ecdhPrivateKey.ECDH(ephemeralECDHPub)
	if err != nil {
		return nil, err
	}

	
	hashedSecret := sha256.Sum256(sharedSecret)

	
	block, err := aes.NewCipher(hashedSecret[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(dataBytes) < 65+nonceSize {
		return nil, fmt.Errorf("invalid encrypted data")
	}

	nonce := dataBytes[65 : 65+nonceSize]
	ciphertext := dataBytes[65+nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
