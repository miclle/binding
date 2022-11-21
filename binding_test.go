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

	type FooBarStruct struct {
		Foo string `json:"foo" form:"foo" xml:"foo"`
		Bar string `json:"bar" form:"bar" xml:"bar"`
	}

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

func requestWithBody(method, path, body string) (req *http.Request) {
	req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
	return
}
