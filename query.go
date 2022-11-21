package binding

import "net/http"

type queryBinding struct{}

func (queryBinding) Bind(req *http.Request, obj any) error {
	values := req.URL.Query()
	return mapForm(obj, values)
}
