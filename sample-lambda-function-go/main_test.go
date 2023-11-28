package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestLambdaHandler(t *testing.T) {
	req := EventRequest{
		Product:  "s3://bucket/path/to/hoge.parquet",
		Customer: "s3://bucket/path/to/fuga.parquet",
		Store:    "s3://bucket/path/to/piyo.parquet",
	}
	response, err := LambdaHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
}

func TestCipher1(t *testing.T) {
	dataKey, err := GetDataKey("/name/of/encryption-key")
	assert.NoError(t, err)

	encryptedStr, err := Encrypt("梅原 大吾", dataKey)
	assert.NoError(t, err)
	// comparing with python result
	assert.Equal(t, "UMtpfOefMSA8HArD78I2YAsWjM+nVAc+FrUCAChmuxo=", encryptedStr)

	decryptedStr, err := Decrypt(encryptedStr, dataKey)
	assert.NoError(t, err)
	assert.Equal(t, "梅原 大吾", decryptedStr)
}

func TestCipher2(t *testing.T) {
	dataKey, err := GetDataKey("/name/of/encryption-key")
	assert.NoError(t, err)

	encryptedStr, err := Encrypt("高柳 光臣", dataKey)
	assert.NoError(t, err)
	// comparing with python result
	assert.Equal(t, "UMtpfOefMSA8HArD78I2YMb/vydARfG5P9htgSiS+YI=", encryptedStr)

	decryptedStr, err := Decrypt(encryptedStr, dataKey)
	assert.NoError(t, err)
	assert.Equal(t, "高柳 光臣", decryptedStr)
}

func TestReadConfig(t *testing.T) {
	expectedMigrationCols := []string{
		"Id",
		"FirstName",
		"LastName",
		"Address",
		"Gender",
		"Birthday",
	}
	expectedEncryptionCols := []string{
		"FirstName",
		"LastName",
		"Address",
	}
	config := NewConfig("./config.toml")
	assert.Equal(t, "/name/of/encryption-key", config.ParamStoreName)
	assert.Equal(t, "glue-connection-name", config.ConnectionName)
	assert.Equal(t, "confidential", config.Customer.Database)
	assert.Equal(t, "pfcs_qua_infs", config.Customer.Table)
	assert.Equal(t, expectedMigrationCols, config.Customer.MigrationCols)
	assert.Equal(t, expectedEncryptionCols, config.Customer.EncryptionCols)

	expectedKeys := []string{"Product", "Customer", "Store"}
	configMap := config.ConvertInfoMap()
	for key, value := range configMap {
		assert.True(t, slices.Contains(expectedKeys, key))
		if key == "Customer" {
			assert.Equal(t, "foo", value.Database)
			assert.Equal(t, "customers", value.Table)
			assert.Equal(t, expectedMigrationCols, value.MigrationCols)
			assert.Equal(t, expectedEncryptionCols, value.EncryptionCols)
		}
	}
}

func TestConnectionHolder(t *testing.T) {
	connHolder := NewConnectionHolder("glue-connection-name")
	assert.NotNil(t, connHolder)

	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.bbbbbbbbbbbb.ap-northeast-1.rds.amazonaws.com:3306", connHolder.Host())
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.bbbbbbbbbbbb.ap-northeast-1.rds.amazonaws.com", connHolder.HostName())
}
