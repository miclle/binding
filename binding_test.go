package binding

import (
	"bytes"
	stdJson "encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var b = &binder{}

type FooStruct struct {
	Foo string `json:"foo" form:"foo" xml:"foo"`
}

type FooBarStruct struct {
	FooStruct
	Bar string `json:"bar" form:"bar" xml:"bar"`
}

type FooStructForMapType struct {
	MapFoo map[string]any `form:"map_foo"`
}

type FooStructForBoolType struct {
	BoolFoo bool `form:"bool_foo"`
}

func TestBindingJSONNilBody(t *testing.T) {
	var obj FooStruct
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "", obj.Foo)
}

func TestBindingJSON(t *testing.T) {
	var obj = FooStruct{}
	req := requestWithBody("POST", "/", `{"foo": "bar"}`)
	req.Header.Set("Content-Type", "application/json")

	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStruct{}
	req = requestWithBody("POST", "/", `{"bar": "foo"}`)
	req.Header.Set("Content-Type", "application/json")

	err = b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.NotEqual(t, "bar", obj.Foo)
	assert.Equal(t, "", obj.Foo)
}

func TestBindingJSONUseNumber(t *testing.T) {

	type FooStructUseNumber struct {
		Foo any `json:"foo"`
	}

	obj := FooStructUseNumber{}
	req := requestWithBody("POST", "/", `{"foo": 123}`)
	req.Header.Set("Content-Type", "application/json")

	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, float64(123), obj.Foo)

	EnableDecoderUseNumber = true

	obj = FooStructUseNumber{}
	req = requestWithBody("POST", "/", `{"foo": 123}`)
	req.Header.Set("Content-Type", "application/json")

	err = b.Bind(req, &obj)
	assert.NoError(t, err)

	v, e := obj.Foo.(stdJson.Number).Int64()
	assert.NoError(t, e)
	assert.Equal(t, int64(123), v)

	obj = FooStructUseNumber{}
	req = requestWithBody("POST", "/", `{"bar": "foo"}`)
	req.Header.Set("Content-Type", "application/json")

	err = b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Nil(t, obj.Foo)
}

func TestBindingForm(t *testing.T) {

	obj := FooBarStruct{}
	req := requestWithBody("POST", "/", "foo=bar&bar=foo")
	req.Header.Add("Content-Type", MIMEPOSTForm)
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)

	obj = FooBarStruct{}
	req = requestWithBody("POST", "/", "bar2=foo")
	req.Header.Add("Content-Type", MIMEPOSTForm)
	err = b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "", obj.Foo)
	assert.Equal(t, "", obj.Bar)

	obj = FooBarStruct{}
	req = requestWithBody("GET", "/?foo=bar&bar=foo", "")
	req.Header.Add("Content-Type", MIMEPOSTForm)
	err = b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)

	obj = FooBarStruct{}
	req = requestWithBody("GET", "/?bar2=foo", "")
	req.Header.Add("Content-Type", MIMEPOSTForm)
	err = b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "", obj.Foo)
	assert.Equal(t, "", obj.Bar)
}

func TestBindingQuery(t *testing.T) {
	testQueryBinding(t, "POST", "/?foo=bar&bar=foo", "/", "foo=unused", "bar2=foo")
}

func TestBindingQuery2(t *testing.T) {
	testQueryBinding(t, "GET", "/?foo=bar&bar=foo", "/?bar2=foo", "foo=unused", "")
}

func TestBindingQueryFail(t *testing.T) {
	testQueryBindingFail(t, "POST", "/?map_foo=", "/", "map_foo=unused", "bar2=foo")
}

func TestBindingQueryFail2(t *testing.T) {
	testQueryBindingFail(t, "GET", "/?map_foo=", "/?bar2=foo", "map_foo=unused", "")
}

func TestBindingQueryBoolFail(t *testing.T) {
	testQueryBindingBoolFail(t, "GET", "/?bool_foo=fasl", "/?bar2=foo", "bool_foo=unused", "")
}

func TestBindingQueryStringMap(t *testing.T) {
	b := Query

	obj := make(map[string]string)
	req := requestWithBody("GET", "/?foo=bar&hello=world", "")
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Len(t, obj, 2)
	assert.Equal(t, "bar", obj["foo"])
	assert.Equal(t, "world", obj["hello"])

	obj = make(map[string]string)
	req = requestWithBody("GET", "/?foo=bar&foo=2&hello=world", "") // should pick last
	err = b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Len(t, obj, 2)
	assert.Equal(t, "2", obj["foo"])
	assert.Equal(t, "world", obj["hello"])
}

func testQueryBinding(t *testing.T, method, path, badPath, body, badBody string) {
	b := Query

	obj := FooBarStruct{}
	req := requestWithBody(method, path, body)
	if method == "POST" {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
}

func testQueryBindingFail(t *testing.T, method, path, badPath, body, badBody string) {
	b := Query

	obj := FooStructForMapType{}
	req := requestWithBody(method, path, body)
	if method == "POST" {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	assert.Error(t, err)
}

func testQueryBindingBoolFail(t *testing.T, method, path, badPath, body, badBody string) {
	b := Query

	obj := FooStructForBoolType{}
	req := requestWithBody(method, path, body)
	if method == "POST" {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	assert.Error(t, err)
}

func requestWithBody(method, path, body string) (req *http.Request) {
	req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
	return
}
