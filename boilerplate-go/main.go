package main

import (
	"context"
	"reflect"

	"github.com/aws/aws-lambda-go/lambda"
)

type EventResponse struct {
	StatusCode int
	Message    string
}

type EventRequest struct {
	Hoge string
	Fuga string
	Piyo string
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
	return EventResponse{200, "Function Succeeded."}, nil
}

func main() {
	lambda.Start(LambdaHandler)
}
