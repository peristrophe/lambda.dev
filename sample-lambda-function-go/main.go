package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"

	"github.com/aws/aws-lambda-go/lambda"
)

type EventResponse struct {
	StatusCode int
	Message    string
}

type EventRequest struct {
	Product  string
	Customer string
	Store    string
}

func (c *EventRequest) ToMap() map[string]string {
	tv := reflect.TypeOf(*c)
	rv := reflect.ValueOf(*c)
	keysValues := map[string]string{}
	for i := 0; i < rv.NumField(); i++ {
		key := tv.Field(i).Name
		value := rv.Field(i).Interface().(string)
		keysValues[key] = value
	}
	return keysValues
}

func LambdaHandler(ctx context.Context, request EventRequest) (EventResponse, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Handle Request", "Request Payload", request)

	requestMap := request.ToMap()
	config := NewConfig("./config.toml")

	connHolder := NewConnectionHolder(config.ConnectionName)
	if connHolder == nil {
		message := fmt.Sprintf("ConnectionHolder construction failed.")
		return EventResponse{500, message}, fmt.Errorf(message)
	}

	dataKey, err := GetDataKey(config.ParamStoreName)
	if err != nil {
		return EventResponse{500, err.Error()}, err
	}

	for targetSym, targetInfo := range config.ConvertInfoMap() {
		logger.Info("Itaration Start", "Target", targetSym, "Target Inof", targetInfo)

		rReader, err := DownloadAndReadParquet(requestMap[targetSym])
		if err != nil {
			return EventResponse{500, err.Error()}, err
		}

		for rReader.Next() {
			records := rReader.Record()
			rows := Rows(records, targetInfo.MigrationCols)

			// Encrypt sensitive informations
			if len(targetInfo.EncryptionCols) > 0 {
				for i, colName := range targetInfo.MigrationCols {
					if !slices.Contains(targetInfo.EncryptionCols, colName) {
						continue
					}
					for j := 0; j < len(rows); j++ {
						before := fmt.Sprintf("%s", rows[j][i])
						rows[j][i], err = Encrypt(before, dataKey)
						if err != nil {
							return EventResponse{500, err.Error()}, err
						}
					}
				}
			}

			// Upsert dataset
			mysql := NewMySQL(connHolder, targetInfo.Database)
			err = mysql.Upsert(rows, targetInfo.Table, targetInfo.MigrationCols)
			if err != nil {
				return EventResponse{500, err.Error()}, err
			}
			logger.Info(fmt.Sprintf("Upsert completed for %s", targetSym), "Rows", rows)
		}
	}

	return EventResponse{200, "Function Succeeded."}, nil
}

func main() {
	lambda.Start(LambdaHandler)
}
