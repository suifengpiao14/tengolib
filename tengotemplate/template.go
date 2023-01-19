package tengotemplate

import (
	"context"
	"fmt"

	"strings"
	"text/template"

	"bytes"

	"github.com/Masterminds/sprig/v3"
	"github.com/d5/tengo/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/tengolib"
	"github.com/suifengpiao14/tengolib/tengocontext"
	gormLogger "gorm.io/gorm/logger"
)

const (
	LOG_INFO_SQL_TEMPLATE = "LogInfoTemplateSQL"
)

type LogInfoTemplateSQL struct {
	Context context.Context
	SQL     string      `json:"sql"`
	Named   string      `json:"named"`
	Data    interface{} `json:"data"`
	Result  string      `json:"result"`
	Err     error       `json:"error"`
}

func (l LogInfoTemplateSQL) GetName() string {
	return LOG_INFO_SQL_TEMPLATE
}
func (l LogInfoTemplateSQL) Error() error {
	return l.Err
}

const (
	EOF                  = "\n"
	WINDOW_EOF           = "\r\n"
	HTTP_HEAD_BODY_DELIM = EOF + EOF
)

type TemplateOut struct {
	tengo.ObjectImpl
	Out  string                 `json:"out"`
	Data map[string]interface{} `json:"data"`
}

func (to *TemplateOut) TypeName() string {
	return "template-out"
}
func (to *TemplateOut) String() string {
	return to.Out
}

func (to *TemplateOut) ToSQL(args ...tengo.Object) (sqlObj tengo.Object, err error) {
	sqlLogInfo := LogInfoTemplateSQL{}
	defer func() {
		sqlLogInfo.Err = err
		tengolib.SendLogInfo(sqlLogInfo)
	}()
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	ctxObj, ok := args[0].(*tengocontext.TengoContext)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "context",
			Expected: "context.Context",
			Found:    args[0].TypeName(),
		}
	}
	sqlStr, err := ToSQL(to.Out, to.Data)
	if err != nil {
		return nil, err
	}
	sqlLogInfo.Context = ctxObj.Context
	sqlObj = &tengo.String{Value: sqlStr}
	sqlLogInfo.SQL = sqlStr
	return sqlObj, nil
}

//TemplateFuncMap 外部需要增加模板自定义函数时,在初始化模板前,设置该变量即可
var TemplateFuncMap = make([]template.FuncMap, 0)

type TengoTemplate struct {
	tengo.ImmutableMap
	Template *template.Template
	tpl      string
}

func NewTemplate() (t *TengoTemplate) {
	tpl := template.New("").Funcs(TemplatefuncMapSQL).Funcs(sprig.TxtFuncMap())
	for _, fnMap := range TemplateFuncMap {
		tpl = tpl.Funcs(fnMap)
	}
	t = &TengoTemplate{
		ImmutableMap: tengo.ImmutableMap{
			Value: make(map[string]tengo.Object),
		},
		Template: tpl,
	}
	t.Value = map[string]tengo.Object{
		"exec": &tengo.UserFunction{
			Value: t.TengoExec,
		},
	}
	return t
}

func (t *TengoTemplate) AddTpl(name string, s string) (tplNames []string) {
	tmpl := t.Template.Lookup(name)
	if tmpl == nil {
		tmpl = t.Template.New(name)
	}
	template.Must(tmpl.Parse(s)) // 追加
	tmp := template.Must(NewTemplate().Template.Parse(s))
	tplNames = GetTemplateNames(tmp)

	t.tpl = fmt.Sprintf(`%s\n{{define "%s"}}%s{{end}}`, t.tpl, name, s)
	return tplNames
}

func (t *TengoTemplate) TypeName() string {
	return "template"
}
func (t *TengoTemplate) String() string {
	return t.tpl
}

func (t *TengoTemplate) TengoExec(args ...tengo.Object) (tplOut tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	tplName, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "tplName",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	tengoMap, ok := args[1].(*tengo.Map)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "data",
			Expected: "map",
			Found:    args[1].TypeName(),
		}
	}
	volume := &VolumeMap{}
	for k, v := range tengoMap.Value {
		volume.SetValue(k, tengo.ToInterface(v))
	}
	var out string
	out, changedVolume, err := t.Exec(tplName, volume)
	if err != nil {
		return nil, err
	}
	tplOut = &TemplateOut{Out: out, Data: changedVolume.ToMap()}
	return tplOut, nil
}
func (t *TengoTemplate) Exec(tplName string, volume VolumeInterface) (out string, changedVolume VolumeInterface, err error) {
	var b bytes.Buffer
	err = t.Template.ExecuteTemplate(&b, tplName, volume)
	if err != nil {
		err = errors.WithStack(err)
		return "", nil, err
	}
	out = strings.ReplaceAll(b.String(), WINDOW_EOF, EOF)
	out = tengolib.TrimSpaces(out)
	return out, volume, nil
}

func GetTemplateNames(t *template.Template) []string {
	out := make([]string, 0)
	for _, tpl := range t.Templates() {
		name := tpl.Name()
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

//ToSQL 将字符串、数据整合为sql
func ToSQL(named string, data map[string]interface{}) (sql string, err error) {
	statment, arguments, err := sqlx.Named(named, data)
	if err != nil {
		err = errors.WithStack(err)
		return "", err
	}
	sql = gormLogger.ExplainSQL(statment, nil, `'`, arguments...)
	return sql, nil
}
