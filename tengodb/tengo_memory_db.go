package tengodb

import (
	"context"

	"github.com/d5/tengo/v2"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/tengolib/tengocontext"
)

type TengoMemoryDB struct {
	tengo.ImmutableMap
	InOutMap map[string]string
}

func (m *TengoMemoryDB) TypeName() string {
	return "memory_db"
}
func (m *TengoMemoryDB) String() string {
	return ""
}
func (m *TengoMemoryDB) ExecOrQueryContext(ctx context.Context, sql string) (out string, err error) {
	out, ok := m.InOutMap[sql]
	if !ok {
		err = errors.Errorf("not found by sql:%s", sql)
		return "", err
	}
	return out, nil
}

func (m *TengoMemoryDB) TengoExecOrQueryContext(args ...tengo.Object) (ret tengo.Object, err error) {
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
	out, err := m.ExecOrQueryContext(ctx, sql)

	ret = &tengo.String{Value: out}
	return ret, err
}

func (m *TengoMemoryDB) BeginTx(args ...tengo.Object) (ret tengo.Object, err error) {
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
	_ = ctxObj.Context

	dbObjPossible := args[1]
	_, ok = dbObjPossible.(*TengoMemoryDB)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "db",
			Expected: "db",
			Found:    dbObjPossible.TypeName(),
		}
	}
	tx := &tengo.String{Value: "memoryDB"}
	return tx, err
}

func NewTengoMemoryDB(config string) (tengoMemoryDB *TengoMemoryDB, err error) {

	tengoMemoryDB = &TengoMemoryDB{
		ImmutableMap: tengo.ImmutableMap{
			Value: make(map[string]tengo.Object),
		},
	}
	//注入tengo 脚本方法
	methods := map[string]tengo.CallableFunc{
		"execOrQueryContext": tengoMemoryDB.TengoExecOrQueryContext,
		"beginTx":            tengoMemoryDB.BeginTx,
	}

	for key, method := range methods {
		tengoMemoryDB.Value[key] = &tengo.UserFunction{
			Name:  key,
			Value: method,
		}

	}
	return tengoMemoryDB, nil
}
