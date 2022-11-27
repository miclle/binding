package binding

import (
	"errors"
	"net/http"
	"reflect"
)

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              = "application/json"                  // json
	MIMEYAML              = "application/x-yaml"                // yaml
	MIMEXML               = "application/xml"                   // xml
	MIMEXML2              = "text/xml"                          // xml
	MIMEPOSTForm          = "application/x-www-form-urlencoded" // form
	MIMEMultipartPOSTForm = "multipart/form-data"               // form
	MIMEPROTOBUF          = "application/x-protobuf"            // protobuf
	MIMETOML              = "application/toml"                  // toml

	MIMEHTML  = "text/html"
	MIMEPlain = "text/plain"
)

var (
	errCantOnlyBindPointer = errors.New("can only bind pointer")
	errInvalidRequest      = errors.New("invalid request")
)

// Binder describes the interface which needs to be implemented for binding the
// data present in the request such as JSON request body, query parameters or
// the form POST.
type Binder interface {
	Bind(*http.Request, any) error
}

// URIBinder adds BindURI method to Binding. BindUri is similar with Bind,
// but it reads the Params.
type URIBinder interface {
	BindURI(map[string][]string, any) error
}

// These implement the Binding interface and can be used to bind the data
// present in the request to struct instances.
var (
	JSON          Binder    = jsonBinder{}
	XML           Binder    = xmlBinding{}
	YAML          Binder    = yamlBinding{}
	Form          Binder    = formBinder{}
	FormMultipart Binder    = formMultipartBinder{}
	ProtoBuf      Binder    = protobufBinding{}
	TOML          Binder    = tomlBinding{}
	Query         Binder    = queryBinding{}
	Header        Binder    = headerBinding{}
	URI           URIBinder = uriBinding{}
)

var defaultBinder Binder

var binders = map[string]Binder{
	MIMEJSON:              JSON,          // json
	MIMEYAML:              YAML,          // yaml
	MIMEXML:               XML,           // xml
	MIMEXML2:              XML,           // xml
	MIMEMultipartPOSTForm: FormMultipart, // form
	MIMEPOSTForm:          Form,          // form
	MIMEPROTOBUF:          ProtoBuf,      // protobuf
	MIMETOML:              TOML,          // toml
}

type binder struct{}

func (binder *binder) Bind(req *http.Request, obj any, params ...map[string][]string) (err error) {

	vPtr := reflect.ValueOf(obj)

	if vPtr.Kind() != reflect.Ptr {
		return errCantOnlyBindPointer
	}

	// bind request body
	// --------------------------------------------------------------------------
	var contentType = filterFlags(req.Header.Get("Content-Type"))

	if binder, exists := binders[contentType]; exists {
		err = binder.Bind(req, obj)
	} else {
		if defaultBinder != nil {
			err = defaultBinder.Bind(req, obj)
		}
	}
	if err != nil {
		return err
	}

	// bind request query, header and uri
	// --------------------------------------------------------------------------
	vPtr = vPtr.Elem()

	for vPtr.Kind() == reflect.Ptr {
		if vPtr.IsNil() {
			vPtr.Set(reflect.New(vPtr.Type().Elem()))
		}
		vPtr = vPtr.Elem()
	}

	if vPtr.Kind() != reflect.Struct {
		return
	}

	var vType = vPtr.Type()
	var hasQueryField, hasURIField, hasHeaderField bool

	for i := 0; i < vPtr.NumField(); i++ {
		field := vType.Field(i)
		if tag := field.Tag.Get("query"); tag != "" && tag != "-" {
			hasQueryField = true
		}
		if tag := field.Tag.Get("url"); tag != "" && tag != "-" {
			hasURIField = true
		}
		if tag := field.Tag.Get("header"); tag != "" && tag != "-" {
			hasHeaderField = true
		}
	}

	if hasQueryField {
		err = Query.Bind(req, obj)
		if err != nil {
			return err
		}
	}

	if hasURIField && len(params) > 0 {
		err = URI.BindURI(params[0], obj)
		if err != nil {
			return err
		}
	}

	if hasHeaderField {
		err = Header.Bind(req, obj)
		if err != nil {
			return err
		}
	}

	return nil
}
