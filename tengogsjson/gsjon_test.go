package tengogsjson

import (
	"fmt"
	"testing"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/stretchr/testify/require"
)

func TestGetSet(t *testing.T) {
	jsonstr := `
	[{"questionId":"12057","questionName":"全新机(包装盒无破损,配件齐全且原装,可无原 机膜和卡针)","classId":"1","className":"手机"},{"questionId":"12097","questionName":"机身弯曲情况","classId":"3","className":"平板"}]
	`
	jsonObj := &tengo.String{Value: jsonstr}
	pathMapObj := &tengo.String{
		Value: `[{"src":"@this.#.questionId","dst":"items.-1.id"},{"src":"@this.#.questionName","dst":"items.-1.name"},{"src":"@this.#.classId","dst":"items.-1.classId"},{"src":"@this.#.className","dst":"items.-1.className"}]`,
	}
	s := tengo.NewScript([]byte(`
	fmt:=import("fmt")
	out:=gsjson.GetSet(jsonstr,pathMap)
	`))
	s.EnableFileImport(true)
	s.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	err := s.Add("gsjson", GSjson)
	if err != nil {
		require.NoError(t, err)
		return
	}
	if err = s.Add("jsonstr", jsonObj); err != nil {
		require.NoError(t, err)
		return
	}
	if err = s.Add("pathMap", pathMapObj); err != nil {
		require.NoError(t, err)
		return
	}

	c, err := s.Compile()
	if err != nil {
		require.NoError(t, err)
		return
	}

	if err := c.Run(); err != nil {
		require.NoError(t, err)
		return
	}

	v := c.Get("out")
	fmt.Println(v)

}

func TestGetSetArr(t *testing.T) {
	jsonstr := `
	[{"questionId":"12057","questionName":"全新机(包装盒无破损,配件齐全且原装,可无原 机膜和卡针)","classId":"1","className":"手机"}],
	[{"questionId":"12097","questionName":"机身弯曲情况","classId":"3","className":"平板"}]
	`
	jsonObj := &tengo.String{Value: jsonstr}
	pathMapObj := &tengo.String{
		Value: `[{"src":"@this.#.questionId","dst":"items.-1.id"},{"src":"@this.#.questionName","dst":"items.-1.name"},{"src":"@this.#.classId","dst":"items.-1.classId"},{"src":"@this.#.className","dst":"items.-1.className"}]`,
	}
	s := tengo.NewScript([]byte(`
	fmt:=import("fmt")
	out:=gsjson.GetSet(jsonstr,pathMap)
	`))
	s.EnableFileImport(true)
	s.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	err := s.Add("gsjson", GSjson)
	if err != nil {
		require.NoError(t, err)
		return
	}
	if err = s.Add("jsonstr", jsonObj); err != nil {
		require.NoError(t, err)
		return
	}
	if err = s.Add("pathMap", pathMapObj); err != nil {
		require.NoError(t, err)
		return
	}

	c, err := s.Compile()
	if err != nil {
		require.NoError(t, err)
		return
	}

	if err := c.Run(); err != nil {
		require.NoError(t, err)
		return
	}

	v := c.Get("out")
	fmt.Println(v)

}
