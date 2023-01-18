package tengotemplate

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/rs/xid"
)

const IN_INDEX = "__inIndex"

var TemplatefuncMapSQL = template.FuncMap{
	"zeroTime":      ZeroTime,
	"currentTime":   CurrentTime,
	"permanentTime": PermanentTime,
	"contains":      strings.Contains,
	"newPreComma":   NewPreComma,
	"in":            In,
	//"toCamel":           ToCamel,
	//"toLowerCamel":      ToLowerCamel,
	//"snakeCase":         SnakeCase,
	//"joinAll":           JoinAll,
	"md5lower":        MD5LOWER,
	"fen2yuan":        Fen2yuan,
	"timestampSecond": TimestampSecond,
	"xid":             Xid,
	//"jsonCompact":       JsonCompact,
	//"standardizeSpaces": util.StandardizeSpaces,
	//"column2Row":        util.Column2Row,
	//"row2Column":        util.Row2Column,
}

func ZeroTime(volume VolumeInterface) (string, error) {
	named := "ZeroTime"
	placeholder := ":" + named
	value := "0000-00-00 00:00:00"
	volume.SetValue(named, value)
	return placeholder, nil
}

func CurrentTime(volume VolumeInterface) (string, error) {
	named := "CurrentTime"
	placeholder := ":" + named
	value := time.Now().Format("2006-01-02 15:04:05")
	volume.SetValue(named, value)
	return placeholder, nil
}

func PermanentTime(volume VolumeInterface) (string, error) {
	named := "PermanentTime"
	placeholder := ":" + named
	value := "3000-12-31 23:59:59"
	volume.SetValue(named, value)
	return placeholder, nil
}

func MD5LOWER(s ...string) string {
	allStr := strings.Join(s, "")
	h := md5.New()
	h.Write([]byte(allStr))
	return hex.EncodeToString(h.Sum(nil))
}

func Fen2yuan(fen interface{}) string {
	var yuan float64
	intFen, ok := fen.(int)
	if ok {
		yuan = float64(intFen) / 100
		return strconv.FormatFloat(yuan, 'f', 2, 64)
	}
	strFen, ok := fen.(string)
	if ok {
		intFen, err := strconv.Atoi(strFen)
		if err == nil {
			yuan = float64(intFen) / 100
			return strconv.FormatFloat(yuan, 'f', 2, 64)
		}
	}
	return strFen
}

// 秒计数的时间戳
func TimestampSecond() int64 {
	return time.Now().Unix()
}

func Xid() string {
	guid := xid.New()
	return guid.String()
}

type preComma struct {
	comma string
}

func NewPreComma() *preComma {
	return &preComma{}
}

func (c *preComma) PreComma() string {
	out := c.comma
	c.comma = ","
	return out
}

func In(volume VolumeInterface, data interface{}) (str string, err error) {
	placeholders := make([]string, 0)
	inIndexKey := IN_INDEX
	var inIndex int
	ok := volume.GetValue(inIndexKey, &inIndex)
	if !ok {
		inIndex = 0
	}

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		num := v.Len()
		for i := 0; i < num; i++ {
			inIndex++
			named := fmt.Sprintf("in_%d", inIndex)
			placeholder := ":" + named
			placeholders = append(placeholders, placeholder)
			volume.SetValue(named, v.Index(i).Interface())
		}

	case reflect.String:
		arr := strings.Split(v.String(), ",")
		num := len(arr)
		for i := 0; i < num; i++ {
			inIndex++
			named := fmt.Sprintf("in_%d", inIndex)
			placeholder := ":" + named
			placeholders = append(placeholders, placeholder)
			volume.SetValue(named, arr[i])
		}
	default:
		err = fmt.Errorf("want slice/array/string ,have %s", v.Kind().String())
		if err != nil {
			return "", err
		}
	}
	volume.SetValue(inIndexKey, inIndex) // 更新InIndex_
	str = strings.Join(placeholders, ",")
	return str, nil

}
