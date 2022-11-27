package binding

import (
	"bytes"
	stdJson "encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/miclle/binding/testdata/protoexample"
)

var b = &binder{}

type FooStruct struct {
	Foo string `json:"foo" yaml:"foo" form:"foo" xml:"foo" toml:"foo" query:"foo"`
}

type FooBarStruct struct {
	FooStruct
	Bar string `json:"bar" yaml:"bar" form:"bar" xml:"bar" toml:"bar" query:"bar"`
}

type FooStructForMapType struct {
	MapFoo map[string]any `form:"map_foo" query:"map_foo"`
}

type FooStructForBoolType struct {
	BoolFoo bool `form:"bool_foo" query:"bool_foo"`
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
	testBodyBinding(t, b, "application/json", "/", "/", `{"foo": "bar"}`, `{"bar": "foo"}`)
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

func TestBindingJSONStringMap(t *testing.T) {
	testBodyBindingStringMap(t, b, "application/json", "/", "/", `{"foo": "bar", "hello": "world"}`, `{"num": 2}`)
}

func TestBindingJSONDisallowUnknownFields(t *testing.T) {
	type FooStructDisallowUnknownFields struct {
		Foo any `json:"foo"`
	}

	EnableDecoderDisallowUnknownFields = true
	defer func() {
		EnableDecoderDisallowUnknownFields = false
	}()

	obj := FooStructDisallowUnknownFields{}
	req := requestWithBody("POST", "/", `{"foo": "bar"}`)
	req.Header.Set("Content-Type", "application/json")
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStructDisallowUnknownFields{}
	req = requestWithBody("POST", "/", `{"foo": "bar", "what": "this"}`)
	req.Header.Set("Content-Type", "application/json")
	err = b.Bind(req, &obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "what")
}

func TestBindingXML(t *testing.T) {
	testBodyBinding(t, b, "application/xml", "/", "/", "<map><foo>bar</foo></map>", "<map><bar>foo</bar></map>")
	testBodyBinding(t, b, "text/xml", "/", "/", "<map><foo>bar</foo></map>", "<map><bar>foo</bar></map>")
}

func TestBindingXMLFail(t *testing.T) {
	testBodyBindingFail(t, b, "application/xml", "/", "/", "<map><foo>bar<foo></map>", "<map><bar>foo</bar></map>")
	testBodyBindingFail(t, b, "text/xml", "/", "/", "<map><foo>bar<foo></map>", "<map><bar>foo</bar></map>")
}

func TestBindingYAML(t *testing.T) {
	testBodyBinding(t, b, "application/x-yaml", "/", "/", `foo: bar`, `bar: foo`)
}

func TestBindingYAMLStringMap(t *testing.T) {
	// YAML is a superset of JSON, so the test below is JSON (to avoid newlines)
	testBodyBindingStringMap(t, b, "application/x-yaml", "/", "/", `{"foo": "bar", "hello": "world"}`, `{"nested": {"foo": "bar"}}`)
}

func TestBindingYAMLFail(t *testing.T) {
	testBodyBindingFail(t, b, "application/x-yaml", "/", "/", `foo:\nbar`, `bar: foo`)
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

func TestBindingProtoBuf(t *testing.T) {
	test := &protoexample.Test{
		Label: proto.String("yes"),
	}
	data, _ := proto.Marshal(test)

	testProtoBodyBinding(t,
		b, "protobuf",
		"/", "/",
		string(data), string(data[1:]))
}

func TestBindingProtoBufFail(t *testing.T) {
	test := &protoexample.Test{
		Label: proto.String("yes"),
	}
	data, _ := proto.Marshal(test)

	testProtoBodyBindingFail(t,
		b, "protobuf",
		"/", "/",
		string(data), string(data[1:]))
}

func TestBindingTOML(t *testing.T) {
	testBodyBinding(t, b, "application/toml", "/", "/", `foo="bar"`, `bar="foo"`)
}

func TestBindingTOMLFail(t *testing.T) {
	testBodyBindingFail(t, b, "application/toml", "/", "/", `foo=\n"bar"`, `bar="foo"`)
}

func TestBindingQuery(t *testing.T) {
	testQueryBinding(t, "POST", "/?foo=bar&bar=foo", "/", "foo=unused", "bar2=foo")
	testQueryBinding(t, "GET", "/?foo=bar&bar=foo", "/?bar2=foo", "foo=unused", "")
}

func TestBindingQueryFail(t *testing.T) {
	testQueryBindingFail(t, "POST", "/?map_foo=", "/", "map_foo=unused", "bar2=foo")
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

func TestHeaderBinding(t *testing.T) {

	type tHeader struct {
		Limit int `header:"limit"`
	}

	var theader tHeader

	req := requestWithBody("GET", "/", "")
	req.Header.Add("limit", "1000")

	assert.NoError(t, b.Bind(req, &theader))
	assert.Equal(t, 1000, theader.Limit)

	req = requestWithBody("GET", "/", "")
	req.Header.Add("fail", `{fail:fail}`)

	type failStruct struct {
		Fail map[string]any `header:"fail"`
	}

	err := b.Bind(req, &failStruct{})
	assert.Error(t, err)
}

func TestURIBinding(t *testing.T) {

	type Tag struct {
		Name string `uri:"name"`
	}
	var tag Tag
	m := make(map[string][]string)
	m["name"] = []string{"thinkerou"}
	assert.NoError(t, URI.BindURI(m, &tag))
	assert.Equal(t, "thinkerou", tag.Name)

	type NotSupportStruct struct {
		Name map[string]any `uri:"name"`
	}
	var not NotSupportStruct
	assert.Error(t, URI.BindURI(m, &not))
	assert.Equal(t, map[string]any{}, not.Name)
}

func TestURIInnerBinding(t *testing.T) {
	type Tag struct {
		Name string `uri:"name"`
		S    struct {
			Age int `uri:"age"`
		}
	}

	expectedName := "mike"
	expectedAge := 25

	m := map[string][]string{
		"name": {expectedName},
		"age":  {strconv.Itoa(expectedAge)},
	}

	var tag Tag
	assert.NoError(t, URI.BindURI(m, &tag))
	assert.Equal(t, tag.Name, expectedName)
	assert.Equal(t, tag.S.Age, expectedAge)
}

func testQueryBinding(t *testing.T, method, path, badPath, body, badBody string) {
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
	obj := FooStructForMapType{}
	req := requestWithBody(method, path, body)
	if method == "POST" {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	assert.Error(t, err)
}

func testQueryBindingBoolFail(t *testing.T, method, path, badPath, body, badBody string) {
	obj := FooStructForBoolType{}
	req := requestWithBody(method, path, body)
	if method == "POST" {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	assert.Error(t, err)
}

func testBodyBinding(t *testing.T, b *binder, contentType, path, badPath, body, badBody string) {
	obj := FooStruct{}
	req := requestWithBody("POST", path, body)
	req.Header.Set("Content-Type", contentType)
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStruct{}
	req = requestWithBody("POST", badPath, badBody)
	req.Header.Set("Content-Type", contentType)
	err = b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "", obj.Foo)
}

func testBodyBindingFail(t *testing.T, b *binder, contentType, path, badPath, body, badBody string) {
	obj := FooStruct{}
	req := requestWithBody("POST", path, body)
	req.Header.Set("Content-Type", contentType)
	err := b.Bind(req, &obj)
	assert.Error(t, err)
	assert.Equal(t, "", obj.Foo)

	obj = FooStruct{}
	req = requestWithBody("POST", badPath, badBody)
	req.Header.Set("Content-Type", contentType)
	err = b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "", obj.Foo)
}

func testBodyBindingStringMap(t *testing.T, b *binder, contentType, path, badPath, body, badBody string) {
	obj := make(map[string]string)
	req := requestWithBody("POST", path, body)
	req.Header.Set("Content-Type", contentType)
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Len(t, obj, 2)
	assert.Equal(t, "bar", obj["foo"])
	assert.Equal(t, "world", obj["hello"])

	if badPath != "" && badBody != "" {
		obj = make(map[string]string)
		req = requestWithBody("POST", badPath, badBody)
		req.Header.Set("Content-Type", contentType)
		err = b.Bind(req, &obj)
		assert.Error(t, err)
	}

	objInt := make(map[string]int)
	req = requestWithBody("POST", path, body)
	req.Header.Set("Content-Type", contentType)
	err = b.Bind(req, &objInt)
	assert.Error(t, err)
}

func requestWithBody(method, path, body string) (req *http.Request) {
	req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
	return
}

func testProtoBodyBinding(t *testing.T, b *binder, name, path, badPath, body, badBody string) {
	obj := protoexample.Test{}
	req := requestWithBody("POST", path, body)
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "yes", *obj.Label)

	obj = protoexample.Test{}
	req = requestWithBody("POST", badPath, badBody)
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err = ProtoBuf.Bind(req, &obj)
	assert.Error(t, err)
}

type hook struct{}

func (h hook) Read([]byte) (int, error) {
	return 0, errors.New("error")
}

func testProtoBodyBindingFail(t *testing.T, b *binder, name, path, badPath, body, badBody string) {
	obj := protoexample.Test{}
	req := requestWithBody("POST", path, body)

	req.Body = io.NopCloser(&hook{})
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err := b.Bind(req, &obj)
	assert.Error(t, err)

	invalidobj := FooStruct{}
	req.Body = io.NopCloser(strings.NewReader(`{"msg":"hello"}`))
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err = b.Bind(req, &invalidobj)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "obj is not ProtoMessage")

	obj = protoexample.Test{}
	req = requestWithBody("POST", badPath, badBody)
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err = ProtoBuf.Bind(req, &obj)
	assert.Error(t, err)
}
