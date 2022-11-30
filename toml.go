package binding

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pelletier/go-toml/v2"
)

type tomlBinding struct{}

func decodeToml(r io.Reader, obj interface{}) error {
	decoder := toml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return decoder.Decode(obj)
}

func (tomlBinding) Bind(req *http.Request, obj interface{}) error {
	return decodeToml(req.Body, obj)
}

func (tomlBinding) BindBody(body []byte, obj interface{}) error {
	return decodeToml(bytes.NewReader(body), obj)
}
