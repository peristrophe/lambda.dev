package main

import (
	"reflect"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ParamStoreName string `toml:"parameter_store_name"`
	ConnectionName string `toml:"connection_name"`
	Product        MigrationTarget
	Customer       MigrationTarget
	Store          MigrationTarget
}

type MigrationTarget struct {
	Database       string   `toml:"database"`
	Table          string   `toml:"table"`
	MigrationCols  []string `toml:"migration_columns"`
	EncryptionCols []string `toml:"encryption_columns"`
}

func NewConfig(filePath string) Config {
	var config Config
	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func (c *Config) Keys() []string {
	tv := reflect.TypeOf(*c)
	var keys []string
	for i := 0; i < tv.NumField(); i++ {
		f := tv.Field(i)
		keys = append(keys, f.Name)
	}
	return keys
}

func (c *Config) ConvertInfoMap() map[string]MigrationTarget {
	tv := reflect.TypeOf(*c)
	rv := reflect.ValueOf(*c)
	keysValues := map[string]MigrationTarget{}
	for i := 0; i < rv.NumField(); i++ {
		if tv.Field(i).Type.String() != "main.MigrationTarget" {
			continue
		}
		key := tv.Field(i).Name
		value := rv.Field(i).Interface().(MigrationTarget)
		keysValues[key] = value
	}
	return keysValues
}
