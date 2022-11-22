package binding

import (
	"errors"
	"net/http"
)

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEMSGPACK           = "application/x-msgpack"
	MIMEMSGPACK2          = "application/msgpack"
	MIMEYAML              = "application/x-yaml"
	MIMETOML              = "application/toml"
)

var (
	errInvalidRequest = errors.New("invalid request")
)

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
	URI                  = uriBinding{}
)

type binder struct{}

func (binder *binder) Bind(req *http.Request, obj any) error {

	var (
		contentType = filterFlags(req.Header.Get("Content-Type"))
		err         error
	)

	switch contentType {
	case MIMEJSON:
		err = JSON.Bind(req, obj)
	case MIMEXML, MIMEXML2:
		err = XML.Bind(req, obj)
	case MIMEYAML:
		err = YAML.Bind(req, obj)
	case MIMEMultipartPOSTForm:
		err = FormMultipart.Bind(req, obj)
	default: // case MIMEPOSTForm:
		err = Form.Bind(req, obj)
	}

	return err
}
