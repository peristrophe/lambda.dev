package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func GetDataKey(parameterStoreName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	ssmClient := ssm.NewFromConfig(cfg)
	response, err := ssmClient.GetParameter(
		context.Background(),
		&ssm.GetParameterInput{
			Name:           aws.String(parameterStoreName),
			WithDecryption: aws.Bool(true),
		},
	)
	if err != nil {
		return "", err
	}

	cipherTextBlob, err := base64.StdEncoding.DecodeString(*response.Parameter.Value)
	if err != nil {
		return "", err
	}
	kmsClient := kms.NewFromConfig(cfg)
	response2, err := kmsClient.Decrypt(context.Background(), &kms.DecryptInput{CiphertextBlob: cipherTextBlob})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(response2.Plaintext), nil
}

func GetFixedIV() ([]byte, error) {
	iv, err := base64.StdEncoding.DecodeString("UMtpfOefMSA8HArD78I2YA==")
	if err != nil {
		nil, err
	}
	return iv, nil
}

func GetRandomIV() ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	return iv, nil
}

func Pkcs7Pad(data []byte) []byte {
	length := aes.BlockSize - (len(data) % aes.BlockSize)
	trailing := bytes.Repeat([]byte{byte(length)}, length)
	return append(data, trailing...)
}

func Pkcs7Unpad(data []byte) []byte {
	dataLength := len(data)
	padLength := int(data[dataLength-1])
	return data[:dataLength-padLength]
}

func Encrypt(plainText string, dataKey string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(dataKey)
	if err != nil {
		return "", err
	}

	data := []byte(plainText)

	iv, err := GetRandomIV()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	padded := Pkcs7Pad(data)
	encryptedBytes := make([]byte, len(padded))
	cbcEncrypter := cipher.NewCBCEncrypter(block, iv)
	cbcEncrypter.CryptBlocks(encryptedBytes, padded)

	encrypted := append(iv, encryptedBytes...)
	encryptedStr := base64.StdEncoding.EncodeToString(encrypted)
	return encryptedStr, nil
}

func Decrypt(encryptedStr string, dataKey string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(dataKey)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := data[:aes.BlockSize]
	encryptedBytes := data[aes.BlockSize:]

	decryptedBytes := make([]byte, len(encryptedBytes))
	cbcDecrypter := cipher.NewCBCDecrypter(block, iv)
	cbcDecrypter.CryptBlocks(decryptedBytes, encryptedBytes)

	decrypted := Pkcs7Unpad(decryptedBytes)
	return string(decrypted), nil
}
