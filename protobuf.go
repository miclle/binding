package binding

import (
	"errors"
	"io/ioutil"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type protobufBinding struct{}

func (b protobufBinding) Bind(req *http.Request, obj any) error {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return b.BindBody(buf, obj)
}

func (protobufBinding) BindBody(body []byte, obj any) error {
	msg, ok := obj.(proto.Message)
	if !ok {
		return errors.New("obj is not ProtoMessage")
	}
	return proto.Unmarshal(body, msg)
}
