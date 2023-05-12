package tengodb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/suifengpiao14/logchan/v2"
	"github.com/suifengpiao14/tengolib/tengotemplate"
	"github.com/suifengpiao14/tengolib/util"
)

type DBConfig struct {
	DSN string `json:"dsn"`
}

type LogName string

func (l LogName) String() string {
	return string(l)
}

type LogInfoEXECSQL struct {
	Context      context.Context
	SQL          string    `json:"sql"`
	Result       string    `json:"result"`
	Err          error     `json:"error"`
	BeginAt      time.Time `json:"beginAt"`
	EndAt        time.Time `json:"endAt"`
	Duration     string    `json:"time"`
	AffectedRows int64     `json:"affectedRows"`
}

func (l LogInfoEXECSQL) GetName() logchan.LogName {
	return LOG_INFO_EXEC_SQL
}
func (l LogInfoEXECSQL) Error() error {
	return l.Err
}

const (
	LOG_INFO_EXEC_SQL LogName = "LogInfoEXECSQL"
)

var DriverName = "mysql"

const (
	SQL_TYPE_SELECT = "SELECT"
	SQL_TYPE_OTHER  = "OTHER"
)

type any = interface{}
type ExectorInterface interface {
	ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error)
	QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error)
}

func ExecOrQueryContext(ctx context.Context, exetor ExectorInterface, sqls string) (out string, err error) {
	sqlLogInfo := LogInfoEXECSQL{}
	defer func() {
		sqlLogInfo.Err = err
		duration := float64(sqlLogInfo.EndAt.Sub(sqlLogInfo.BeginAt).Nanoseconds()) / 1e6
		sqlLogInfo.Duration = fmt.Sprintf("%.3fms", duration)
		logchan.SendLogInfo(sqlLogInfo)
	}()
	sqls = util.StandardizeSpaces(util.TrimSpaces(sqls)) // 格式化sql语句
	sqlLogInfo.SQL = sqls
	sqlType := SQLType(sqls)
	if sqlType != SQL_TYPE_SELECT {
		sqlLogInfo.BeginAt = time.Now().Local()
		res, err := exetor.ExecContext(ctx, sqls)
		if err != nil {
			return "", err
		}
		sqlLogInfo.EndAt = time.Now().Local()
		sqlLogInfo.AffectedRows, _ = res.RowsAffected()
		lastInsertId, _ := res.LastInsertId()
		if lastInsertId > 0 {
			return strconv.FormatInt(lastInsertId, 10), nil
		}
		rowsAffected, _ := res.RowsAffected()
		return strconv.FormatInt(rowsAffected, 10), nil
	}
	sqlLogInfo.BeginAt = time.Now().Local()
	rows, err := exetor.QueryContext(ctx, sqls)
	sqlLogInfo.EndAt = time.Now().Local()
	if err != nil {
		return "", err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}()
	allResult := make([][]map[string]string, 0)
	rowsAffected := 0
	for {
		records := make([]map[string]string, 0)
		for rows.Next() {
			rowsAffected++
			var record = make(map[string]interface{})
			var recordStr = make(map[string]string)
			err := MapScan(*rows, record)
			if err != nil {
				return "", err
			}
			for k, v := range record {
				if v == nil {
					recordStr[k] = ""
				} else {
					recordStr[k] = fmt.Sprintf("%s", v)
				}
			}
			records = append(records, recordStr)
		}
		allResult = append(allResult, records)
		if !rows.NextResultSet() {
			break
		}
	}
	sqlLogInfo.AffectedRows = int64(rowsAffected)

	if len(allResult) == 1 { // allResult 初始值为[[]],至少有一个元素
		result := allResult[0]
		if len(result) == 0 { // 结果为空，返回空字符串
			return "", nil
		}
		if len(result) == 1 && len(result[0]) == 1 {
			row := result[0]
			for _, val := range row {
				return val, nil // 只有一个值时，直接返回值本身
			}
		}
		jsonByte, err := json.Marshal(result)
		if err != nil {
			return "", err
		}
		out = string(jsonByte)
		sqlLogInfo.Result = out
		return out, nil
	}

	jsonByte, err := json.Marshal(allResult)
	if err != nil {
		return "", err
	}
	out = string(jsonByte)
	sqlLogInfo.Result = out
	return out, nil
}

// MapScan copy sqlx
func MapScan(r sql.Rows, dest map[string]interface{}) error {
	// ignore r.started, since we needn't use reflect for anything.
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err = r.Scan(values...)
	if err != nil {
		return err
	}

	for i, column := range columns {
		dest[column] = *(values[i].(*interface{}))
	}

	return r.Err()
}

// SQLType 判断 sql  属于那种类型
func SQLType(sqls string) string {
	sqlArr := strings.Split(sqls, tengotemplate.EOF)
	selectLen := len(SQL_TYPE_SELECT)
	for _, sql := range sqlArr {
		if len(sql) < selectLen {
			continue
		}
		pre := sql[:selectLen]
		if strings.ToUpper(pre) == SQL_TYPE_SELECT {
			return SQL_TYPE_SELECT
		}
	}
	return SQL_TYPE_OTHER
}
