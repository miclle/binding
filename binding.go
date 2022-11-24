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
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEMSGPACK           = "application/x-msgpack"
	MIMEMSGPACK2          = "application/msgpack"
	MIMETOML              = "application/toml"

	MIMEHTML  = "text/html"
	MIMEPlain = "text/plain"
)

var (
	errCantOnlyBindPointer = errors.New("can only bind pointer")
	errInvalidRequest      = errors.New("invalid request")
)

var binders = map[string]Binder{
	MIMEJSON:              JSON,          // json
	MIMEYAML:              YAML,          // yaml
	MIMEXML:               XML,           // xml
	MIMEXML2:              XML,           // xml
	MIMEMultipartPOSTForm: FormMultipart, // form
	MIMEPOSTForm:          Form,          // form
}

// Binder describes the interface which needs to be implemented for binding the
// data present in the request such as JSON request body, query parameters or
// the form POST.
type Binder interface {
	Bind(*http.Request, any) error
}

// These implement the Binding interface and can be used to bind the data
// present in the request to struct instances.
var (
	JSON          Binder = jsonBinder{}
	XML           Binder = xmlBinding{}
	YAML          Binder = yamlBinding{}
	Form          Binder = formBinder{}
	FormMultipart Binder = formMultipartBinder{}
	Query         Binder = queryBinding{}
	Header        Binder = headerBinding{}
	URI                  = uriBinding{}
)

type binder struct{}

func (binder *binder) Bind(req *http.Request, obj any) (err error) {

	vPtr := reflect.ValueOf(obj)

	if vPtr.Kind() != reflect.Ptr {
		return errCantOnlyBindPointer
	}

	// bind request body
	// --------------------------------------------------------------------------
	var contentType = filterFlags(req.Header.Get("Content-Type"))

	if binder, exists := binders[contentType]; exists {
		err = binder.Bind(req, obj)
		if err != nil {
			return err
		}
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

	if hasURIField {
		// TODO(m)
	}

	if hasHeaderField {
		err = Header.Bind(req, obj)
		if err != nil {
			return err
		}
	}

	return nil
}
