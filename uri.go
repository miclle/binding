package binding

type uriBinding struct{}

func (uriBinding) BindURI(params map[string][]string, obj interface{}) error {
	return mapFormByTag(obj, params, "uri")
}
