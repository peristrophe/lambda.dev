package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Transpose(slice [][]any) [][]any {
	xl := len(slice[0])
	yl := len(slice)
	result := make([][]any, xl)
	for i := range result {
		result[i] = make([]any, yl)
	}
	for i := 0; i < xl; i++ {
		for j := 0; j < yl; j++ {
			result[i][j] = slice[j][i]
		}
	}
	return result
}

func Values(record arrow.Record, columnName string) []any {
	schema := record.Schema()
	jsonBytes, _ := record.Column(schema.FieldIndices(columnName)[0]).MarshalJSON()
	var values []any
	json.Unmarshal(jsonBytes, &values)
	return values
}

func Rows(record arrow.Record, columnNames []string) [][]any {
	var columns [][]any
	for _, colName := range columnNames {
		values := Values(record, colName)
		columns = append(columns, values)
	}
	return Transpose(columns)
}

func DownloadAndReadParquet(s3FullPath string) (pqarrow.RecordReader, error) {
	splited := strings.Split(s3FullPath, "/")
	bucket := splited[2]
	key := strings.Join(splited[3:], "/")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	response, err := client.GetObject(
		context.Background(),
		&s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)},
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	bReader := bytes.NewReader(buf.Bytes())

	pqReader, err := file.NewParquetReader(bReader)
	if err != nil {
		return nil, err
	}
	defer pqReader.Close()

	fReader, err := pqarrow.NewFileReader(
		pqReader,
		pqarrow.ArrowReadProperties{BatchSize: 1000},
		memory.DefaultAllocator,
	)
	if err != nil {
		return nil, err
	}

	rReader, err := fReader.GetRecordReader(context.Background(), nil, nil)
	if err != nil {
		return nil, err
	}

	return rReader, nil
}
