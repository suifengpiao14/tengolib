package tengocontext

import (
	context "context"

	"github.com/d5/tengo/v2"
)

type TengoContext struct {
	tengo.ObjectImpl
	Context context.Context
}

func (c *TengoContext) TypeName() string {
	return "context"
}
func (c *TengoContext) String() string {
	return "context"
}

var Ctx = map[string]tengo.Object{
	"background": &tengo.UserFunction{
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			ret = &TengoContext{
				Context: context.Background(),
			}
			return ret, nil
		},
	},
	"todo": &tengo.UserFunction{
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			ret = &TengoContext{
				Context: context.TODO(),
			}
			return ret, nil
		},
	},
}

//TengoContextCallable 在tengo脚本中获取新的上下文
func TengoContextCallable(args ...tengo.Object) (ret tengo.Object, err error) {
	ret = &TengoContext{
		Context: context.Background(),
	}
	return
}
