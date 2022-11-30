package binding

import (
	"bytes"
	"io"
	"net/http"

	"gopkg.in/yaml.v3"
)

type yamlBinding struct{}

func (yamlBinding) Bind(req *http.Request, obj interface{}) error {
	return decodeYAML(req.Body, obj)
}

func (yamlBinding) BindBody(body []byte, obj interface{}) error {
	return decodeYAML(bytes.NewReader(body), obj)
}

func decodeYAML(r io.Reader, obj interface{}) error {
	decoder := yaml.NewDecoder(r)
	return decoder.Decode(obj)
}
