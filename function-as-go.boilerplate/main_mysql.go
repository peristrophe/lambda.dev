package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	glueTypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/go-sql-driver/mysql"
)

type ConnectionHolder struct {
	connection glueTypes.Connection
	url        url.URL
}

func NewConnectionHolder(connectionName string) *ConnectionHolder {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil
	}

	client := glue.NewFromConfig(cfg)
	connection, err := client.GetConnection(
		context.Background(),
		&glue.GetConnectionInput{Name: aws.String(connectionName)},
	)
	if err != nil {
		return nil
	}

	connURL := connection.Connection.ConnectionProperties["JDBC_CONNECTION_URL"]
	if strings.HasPrefix(connURL, "jdbc:") {
		connURL = strings.Replace(connURL, "jdbc:", "", 1)
	}
	parsedURL, err := url.Parse(connURL)
	if err != nil {
		return nil
	}

	return &ConnectionHolder{*connection.Connection, *parsedURL}
}

func (h *ConnectionHolder) UserName() string {
	return h.connection.ConnectionProperties["USERNAME"]
}

func (h *ConnectionHolder) Password() string {
	return h.connection.ConnectionProperties["PASSWORD"]
}

func (h *ConnectionHolder) URL() string {
	return h.connection.ConnectionProperties["JDBC_CONNECTION_URL"]
}

func (h *ConnectionHolder) HostName() string {
	return h.url.Hostname()
}

func (h *ConnectionHolder) Host() string {
	// contain port
	return h.url.Host
}

type MySQL struct {
	config mysql.Config
}

func NewMySQL(holder *ConnectionHolder, database string) *MySQL {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	mysqlcfg := mysql.Config{
		DBName:               database,
		User:                 holder.UserName(),
		Passwd:               holder.Password(),
		Addr:                 holder.Host(),
		Net:                  "tcp",
		ParseTime:            true,
		Collation:            "utf8mb4_unicode_ci",
		Loc:                  jst,
		AllowNativePasswords: true,
	}
	return &MySQL{config: mysqlcfg}
}

func (m *MySQL) Upsert(dataset [][]any, table string, columns []string) error {
	db, err := sql.Open("mysql", m.config.FormatDSN())
	if err != nil {
		return err
	}

	var lcols []string
	var valuesStmtArr []string
	var updateStmtArr []string
	var valueCollection []any
	for _, column := range columns {
		lcol := strings.ToLower(column)
		lcols = append(lcols, lcol)
		if lcol != "id" {
			updateStmtArr = append(updateStmtArr, fmt.Sprintf("%s = VALUES(%s)", lcol, lcol))
		}
	}
	for _, row := range dataset {
		valueCollection = append(valueCollection, row...)
		tmpArr := []string{}
		for j := 0; j < len(row); j++ {
			tmpArr = append(tmpArr, "?")
		}
		valuesStmtArr = append(valuesStmtArr, fmt.Sprintf("(%s)", strings.Join(tmpArr, ", ")))
	}
	columnStmt := strings.Join(lcols, ", ")
	valuesStmt := strings.Join(valuesStmtArr, ", ")
	updateStmt := strings.Join(updateStmtArr, ", ")

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s",
		table,
		columnStmt,
		valuesStmt,
		updateStmt,
	)

	ups, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	defer ups.Close()

	res, err := ups.Exec(valueCollection...)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}
