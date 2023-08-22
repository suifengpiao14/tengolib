package tengodb

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/d5/tengo/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/tengolib/tengocontext"
)

// TengoDBInterface 为了实现 memory_db 替换,改成接口，对 事务没有实现替换，后续可以扩展memory_db 时，加上事务方法
type TengoDBInterface interface {
	tengo.Object
	ExecOrQueryContext(ctx context.Context, sql string) (out string, err error)
}

type TengoDB struct {
	tengo.ImmutableMap
	sqlDB *sql.DB
}

func (tengoDB *TengoDB) TypeName() string {
	return "db"
}
func (tengoDB *TengoDB) String() string {
	return ""
}

func (tengoDB *TengoDB) GetDB() (db *sql.DB) {
	return tengoDB.sqlDB
}

var tengoDBMap = make(map[string]*TengoDB)

func NewTengoDB(config string) (tengoDB *TengoDB, err error) {
	tengoDB, ok := tengoDBMap[config]
	if ok {
		return tengoDB, nil
	}
	tengoDB = &TengoDB{
		ImmutableMap: tengo.ImmutableMap{
			Value: make(map[string]tengo.Object),
		},
	}
	cfg := &DBConfig{}
	err = json.Unmarshal([]byte(config), cfg)
	if err != nil {
		return nil, err
	}
	var db *sql.DB

	db, err = sql.Open(DriverName, cfg.DSN)
	if err != nil {
		return // 此处返回闭包,带出error
	}
	tengoDB.sqlDB = db

	if err != nil {
		err = errors.WithMessagef(err, "sql.Open:%s", cfg.DSN)
		return nil, err
	}
	if tengoDB.sqlDB == nil {
		err = errors.New("tengoDB.sqlDB is nil")
		panic(err)
	}
	//注入tengo 脚本方法
	methods := map[string]tengo.CallableFunc{
		"execOrQueryContext": tengoDB.TengoExecOrQueryContext,
		"beginTx":            tengoDB.BeginTx,
	}

	for key, method := range methods {
		tengoDB.Value[key] = &tengo.UserFunction{
			Name:  key,
			Value: method,
		}

	}
	tengoDBMap[config] = tengoDB
	return tengoDB, nil
}

func (db *TengoDB) TengoExecOrQueryContext(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	ctxObjPossible := args[0]
	ctxObj, ok := ctxObjPossible.(*tengocontext.TengoContext)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "context",
			Expected: "context.Context",
			Found:    ctxObjPossible.TypeName(),
		}
	}
	sqlObj := args[1]
	sql, ok := tengo.ToString(sqlObj)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "sql",
			Expected: "string",
			Found:    sqlObj.TypeName(),
		}
	}
	ctx := ctxObj.Context

	out, err := db.ExecOrQueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	ret = &tengo.String{Value: out}
	return ret, err
}

// 简单封装 ExecOrQueryContext, 减少一个参数，可以实现 memory_db 替换，如果直接用方法 ExecOrQueryContext,替换类会非常麻烦
func (db *TengoDB) ExecOrQueryContext(ctx context.Context, sql string) (out string, err error) {
	out, err = ExecOrQueryContext(ctx, db.sqlDB, sql)
	return out, err
}

func (tengoDB *TengoDB) BeginTx(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	ctxObjPossible := args[0]
	ctxObj, ok := ctxObjPossible.(*tengocontext.TengoContext)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "context",
			Expected: "context.Context",
			Found:    ctxObjPossible.TypeName(),
		}
	}
	ctx := ctxObj.Context
	tx, err := newTengoTx(ctx, tengoDB.sqlDB)
	return tx, err
}

type TengoTx struct {
	tengo.ImmutableMap
	sqlTx *sql.Tx
}

func (t *TengoTx) TypeName() string {
	return "db-tx"
}
func (t *TengoTx) String() string {
	return ""
}

func (t *TengoTx) Commit(args ...tengo.Object) (ret tengo.Object, err error) {
	err = t.sqlTx.Commit()
	return nil, err
}

func (t *TengoTx) ExecOrQueryContext(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	ctxObjPossible := args[0]
	ctxObj, ok := ctxObjPossible.(*tengocontext.TengoContext)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "context",
			Expected: "context.Context",
			Found:    ctxObjPossible.TypeName(),
		}
	}
	ctx := ctxObj.Context

	sqlObj := args[1]
	sql, ok := tengo.ToString(sqlObj)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "sql",
			Expected: "string",
			Found:    sqlObj.TypeName(),
		}
	}
	out, err := ExecOrQueryContext(ctx, t.sqlTx, sql)
	if err != nil {
		return nil, err
	}
	ret = &tengo.String{Value: out}
	return ret, err
}
func (t *TengoTx) Rollback(args ...tengo.Object) (ret tengo.Object, err error) {
	err = t.sqlTx.Rollback()
	return nil, err
}

func newTengoTx(ctx context.Context, db *sql.DB) (t *TengoTx, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	t = &TengoTx{
		ImmutableMap: tengo.ImmutableMap{
			Value: make(map[string]tengo.Object),
		},
		sqlTx: tx,
	}

	methods := map[string]tengo.CallableFunc{
		"commit":             t.Commit,
		"execOrQueryContext": t.ExecOrQueryContext,
		"rollback":           t.Rollback,
	}
	for name, fn := range methods {
		t.Value[name] = &tengo.UserFunction{
			Name:  name,
			Value: fn,
		}
	}
	return t, err
}
