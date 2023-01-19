package tengogsjson

import (
	"fmt"
	"strings"

	"github.com/d5/tengo/v2"
	"github.com/pkg/errors"
	_ "github.com/suifengpiao14/gjsonmodifier"
	"github.com/suifengpiao14/tengolib/tengocontext"
	"github.com/suifengpiao14/tengolib/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Storage struct {
	tengo.ImmutableMap
	DiskSpace string
	Memory    tengo.Object //共享内存空间
	Ctx       *tengocontext.TengoContext
}

func NewStorage() (m *Storage) {
	m = &Storage{
		Memory: &tengo.Map{},
		Ctx:    &tengocontext.TengoContext{},
	}
	m.Value = map[string]tengo.Object{
		"Get": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				jsonstr := &tengo.String{Value: m.DiskSpace}
				newArgs := make([]tengo.Object, 0)
				newArgs = append(newArgs, jsonstr)
				newArgs = append(newArgs, args...)
				result, err := Get(newArgs...)
				ret = &tengo.String{Value: result}
				return ret, err
			},
		},
		"Set": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				jsonstr := &tengo.String{Value: m.DiskSpace}
				newArgs := make([]tengo.Object, 0)
				newArgs = append(newArgs, jsonstr)
				newArgs = append(newArgs, args...)
				m.DiskSpace, err = Set(newArgs...)
				return m, err
			},
		},
		"SetRaw": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				jsonstr := &tengo.String{Value: m.DiskSpace}
				newArgs := make([]tengo.Object, 0)
				newArgs = append(newArgs, jsonstr)
				newArgs = append(newArgs, args...)
				m.DiskSpace, err = SetRaw(newArgs...)
				return m, err
			},
		},
		"GetSet": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				jsonstr := &tengo.String{Value: m.DiskSpace}
				newArgs := make([]tengo.Object, 0)
				newArgs = append(newArgs, jsonstr)
				newArgs = append(newArgs, args...)
				m.DiskSpace, err = GetSet(newArgs...)
				return m, err
			},
		},
		"Delete": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				jsonstr := &tengo.String{Value: m.DiskSpace}
				newArgs := make([]tengo.Object, 0)
				newArgs = append(newArgs, jsonstr)
				newArgs = append(newArgs, args...)
				m.DiskSpace, err = Delete(newArgs...)
				return m, err
			},
		},
		"GetMemory": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				return m.Memory, err
			},
		},
		"GetCtx": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				return m.Ctx, err
			},
		},
	}
	return m
}

func (s *Storage) TypeName() string {
	return "gjson-Storage"
}
func (s *Storage) String() string {
	return s.DiskSpace
}

func (s *Storage) CanCall() bool {
	return false
}

var GSjson = map[string]tengo.Object{
	"Get": &tengo.UserFunction{
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			s, err := Get(args...)
			ret = &tengo.String{Value: s}
			return ret, err
		},
	},
	"Set": &tengo.UserFunction{
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			s, err := Set(args...)
			ret = &tengo.String{Value: s}
			return ret, err
		},
	},
	"SetRaw": &tengo.UserFunction{
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			s, err := SetRaw(args...)
			ret = &tengo.String{Value: s}
			return ret, err
		},
	},
	"GetSet": &tengo.UserFunction{
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			s, err := GetSet(args...)
			ret = &tengo.String{Value: s}
			return ret, err
		},
	},
	"Delete": &tengo.UserFunction{
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			s, err := Delete(args...)
			ret = &tengo.String{Value: s}
			return ret, err
		},
	},
}

func Get(args ...tengo.Object) (result string, err error) {
	if len(args) != 2 {
		return "", tengo.ErrWrongNumArguments
	}
	jsonStr, ok := tengo.ToString(args[0])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.get.arg1",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	path, ok := tengo.ToString(args[1])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.get.arg2",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}
	jsonStr = util.TrimSpaces(jsonStr)
	path = util.TrimSpaces(path)
	gResult := gjson.Get(jsonStr, path)
	result = gResult.String()
	return result, nil
}

func Set(args ...tengo.Object) (result string, err error) {
	if len(args) != 3 {
		return "", tengo.ErrWrongNumArguments
	}
	jsonStr, ok := tengo.ToString(args[0])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.get.arg1",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	path, ok := tengo.ToString(args[1])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.get.arg2",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}
	value := tengo.ToInterface(args[2])
	str, err := sjson.Set(jsonStr, path, value)
	if err != nil {
		err = errors.WithMessage(err, "sjson.set")
		return "", err
	}
	return str, nil
}
func SetRaw(args ...tengo.Object) (result string, err error) {
	if len(args) != 3 {
		return "", tengo.ErrWrongNumArguments
	}
	jsonStr, ok := tengo.ToString(args[0])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.get.arg1",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	path, ok := tengo.ToString(args[1])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.get.arg2",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}
	value, ok := tengo.ToString(args[2])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.get.arg3",
			Expected: "string",
			Found:    args[2].TypeName(),
		}
	}
	str, err := sjson.SetRaw(jsonStr, path, value)
	if err != nil {
		err = errors.WithMessage(err, "sjson.setRaw")
		return "", err
	}
	return str, nil
}

func GetSet(args ...tengo.Object) (result string, err error) {
	if len(args) != 2 {
		return "", tengo.ErrWrongNumArguments
	}
	jsonStr, ok := tengo.ToString(args[0])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.GetSet.arg1",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	path, ok := tengo.ToString(args[1])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.GetSet.arg2",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}
	resultArr := gjson.Parse(path).Array()
	var out = ""
	arrayKeys := map[string]struct{}{}
	for _, keyRow := range resultArr {
		src := keyRow.Get("src").String()
		dst := keyRow.Get("dst").String()
		n1Index := strings.LastIndex(dst, "-1")
		if n1Index > -1 {
			parentKey := dst[:n1Index]
			parentKey = strings.TrimRight(parentKey, ".#")
			if parentKey[len(parentKey)-1] == ')' {
				parentKey = fmt.Sprintf("%s#", parentKey)
			}
			arrayKeys[parentKey] = struct{}{}
			dst = fmt.Sprintf("%s%s", dst[:n1Index-1], dst[n1Index+2:])
		}

		raw := gjson.Get(jsonStr, src).Raw
		out, err = sjson.SetRaw(out, dst, raw)
		if err != nil {
			err = errors.WithMessage(err, "gsjson.GetSet")
			return "", err
		}
	}
	for path := range arrayKeys {
		qPath := fmt.Sprintf("%s|@group", path)
		raw := gjson.Get(out, qPath).Raw
		out, err = sjson.SetRaw(out, path, raw)
		if err != nil {
			err = errors.WithMessage(err, "gsjson.GetSet|@group")
			return "", err
		}
	}
	return out, nil
}

func Delete(args ...tengo.Object) (result string, err error) {
	if len(args) != 2 {
		return "", tengo.ErrWrongNumArguments
	}
	jsonStr, ok := tengo.ToString(args[0])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.delete.arg1",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	path, ok := tengo.ToString(args[1])
	if !ok {
		return "", tengo.ErrInvalidArgumentType{
			Name:     "gjson.delete.arg2",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}
	str, err := sjson.Delete(jsonStr, path)
	if err != nil {
		err = errors.WithMessage(err, "sjson.delete")
		return "", err
	}
	return str, nil
}
