package binding

import (
	"errors"
	"io"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type protobufBinding struct{}

func (b protobufBinding) Bind(req *http.Request, obj interface{}) error {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return b.BindBody(buf, obj)
}

func (protobufBinding) BindBody(body []byte, obj interface{}) error {
	msg, ok := obj.(proto.Message)
	if !ok {
		return errors.New("obj is not ProtoMessage")
	}
	return proto.Unmarshal(body, msg)
}
