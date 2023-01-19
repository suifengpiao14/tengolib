package tengolib

import (
	"fmt"
	"testing"

	"github.com/d5/tengo/v2"
)

func TestTengocollection(t *testing.T) {
	script := tengo.NewScript([]byte(`
	collection:=import("collection")
	records:=[
		{"id":1,"name":"张三","age":20,"sex":"男"},
		{"id":2,"name":"李四","age":20,"sex":"男"},
		{"id":3,"name":"王五","age":18,"sex":"女"}
		]
		column:=collection.Column(records,"sex","name")
		index:=collection.Index(records,"id")
		map:=collection.Map(records,func(record,i){
			record["status"]=1
			return record
		})
		group:=collection.Group(records,"sex")
		keyConvert:=collection.KeyConvert(records,{"id":"userId","name":"userName"})
		orderBy:=collection.OrderBy(records,func(r1,r2){
			return r1.id>r2.id
		})
`))
	script.SetImports(GetModuleMap(AllModuleNames()...))
	c, err := script.Run()
	if err != nil {
		panic(err)
	}
	v := c.Get("column")
	fmt.Println(v)
	v = c.Get("index")
	fmt.Println(v)
	v = c.Get("map")
	fmt.Println(v)
	v = c.Get("group")
	fmt.Println(v)
	v = c.Get("keyConvert")
	fmt.Println(v)
	v = c.Get("orderBy")
	fmt.Println(v)
}
