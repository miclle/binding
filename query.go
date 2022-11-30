package binding

import "net/http"

type queryBinding struct{}

func (queryBinding) Bind(req *http.Request, obj interface{}) error {
	values := req.URL.Query()
	return mapFormWithTag(obj, values, "query")
}
