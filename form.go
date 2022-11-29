package binding

import (
	"errors"
	"net/http"
)

const defaultMemory = 32 << 20

type formBinder struct{}

func (formBinder) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := req.ParseMultipartForm(defaultMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return err
	}
	return mapForm(obj, req.Form)
}

type formMultipartBinder struct{}

func (formMultipartBinder) Bind(req *http.Request, obj any) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		return err
	}
	return mappingByPtr(obj, (*multipartRequest)(req), "form")
}
