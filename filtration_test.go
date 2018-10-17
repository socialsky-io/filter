package filter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFiltration(t *testing.T) {
	is := assert.New(t)

	fl := New(map[string]interface{}{
		"key0": " abc ",
		"key1": "2",
		"sub":  map[string]string{"k0": "v0"},
	})

	is.Equal("strToTime", Name("str2time"))
	is.Equal("some", Name("some"))

	is.Equal("", fl.String("key0"))
	is.Equal(0, fl.Int("key1"))

	val, ok := fl.Get("key1")
	is.True(ok)
	is.Equal("2", val)

	val, ok = fl.Get("sub.k0")
	is.True(ok)
	is.Equal("v0", val)

	val, ok = fl.Safe("key1")
	is.False(ok)
	is.Equal(nil, val)

	val, ok = fl.Raw("key1")
	is.True(ok)
	is.Equal("2", val)
	val, ok = fl.Raw("not-exist")
	is.False(ok)
	is.Equal(nil, val)

	f := New(map[string]interface{}{
		"name":  " inhere ",
		"email": " my@email.com ",
	})
	f.AddRules(map[string]string{
		"email": "email",
		"name":  "trim|ucFirst",
	})

	is.Nil(f.Sanitize())
	is.Equal("Inhere", f.String("name"))
	is.Equal("my@email.com", f.String("email"))
}

func TestFiltration_Filtering(t *testing.T) {
	is := assert.New(t)

	data := map[string]interface{}{
		"name":     "inhere",
		"age":      "50",
		"money":    "50.34",
		"remember": "yes",
		//
		"sub":  map[string]string{"k0": "v0"},
		"sub1": []string{"1", "2"},
		"tags": "go;lib",
		"str1": " word ",
		"ids":  []int{1, 2, 2, 1},
	}
	f := New(data)
	f.AddRule("name", "upper")
	f.AddRule("age", "int")
	f.AddRule("money", "float")
	f.AddRule("remember", "bool")
	f.AddRule("sub1", "strings2ints")
	f.AddRule("tags", "str2arr:;")
	f.AddRule("ids", "unique")
	f.AddRule("str1", "ltrim|rtrim")
	f.AddRule("not-exist", "unique")

	is.Nil(f.Filtering())
	is.Nil(f.Filtering())
	is.True(f.IsOK())

	// get value
	is.True(f.Bool("remember"))
	is.False(f.Bool("not-exist"))
	is.Equal(50, f.Int("age"))
	is.Equal(0, f.Int("not-exist"))
	is.Equal(50, f.MustGet("age"))
	is.Equal(int64(50), f.Int64("age"))
	is.Equal(int64(0), f.Int64("not-exist"))
	is.Equal(50.34, f.MustGet("money"))
	is.Equal([]int{1, 2}, f.MustGet("sub1"))
	is.Len(f.MustGet("ids"), 2)
	is.Equal([]string{"go", "lib"}, f.MustGet("tags"))
	is.Equal("INHERE", f.FilteredData()["name"])
	is.Equal("word", f.String("str1"))

	f = New(data)
	f.AddRule("name", "int")
	is.Error(f.Sanitize())

	data["name"] = " inhere "
	data["sDate"] = "2018-10-16 12:34"
	data["msg"] = " hello world "
	data["msg1"] = "helloWorld"
	data["msg2"] = "hello_world"
	f = New(data)
	f.AddRules(map[string]string{
		"age":   "uint",
		"money": "float",
		"name":  "trim|ucFirst",
		"str1":  "trim|upper",
		"sDate": "str2time",
		"msg":   "trim|ucWord",
		"msg1":  "snake",
		"msg2":  "camel",
	})
	is.Nil(f.Sanitize())
	is.Equal("Inhere", f.String("name"))
	is.Equal("WORD", f.String("str1"))
	is.Equal("Hello World", f.String("msg"))
	is.Equal("hello_world", f.String("msg1"))
	is.Equal("helloWorld", f.String("msg2"))

	sTime, ok := f.Safe("sDate")
	is.True(ok)
	is.Equal("2018-10-16 12:34:00 +0000 UTC", fmt.Sprintf("%v", sTime))

	data["url"] = "a.com?p=1"
	f = New(data)
	f.AddRule("url", "urlEncode")
	f.AddRule("msg1", "substr:0,2")
	is.Nil(f.Sanitize())
	is.Equal("he", f.String("msg1"))
	is.Equal("a.com?p%3D1", f.String("url"))
}