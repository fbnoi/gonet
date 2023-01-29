package binding

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var DefaultValidator = &Validator{
	Validate: validator.New(),
}

var (
	JSON          = &jsonBinding{}
	Query         = &queryBinding{}
	Form          = &formBinding{}
	FormMultipart = &formMultipartBinding{}
	FormPost      = &formPostBinding{}
)

const (
	MIME_JSON              = "application/json"
	MIME_HTML              = "text/html"
	MIME_XML               = "application/xml"
	MIME_Plain             = "text/plain"
	MIME_POSTForm          = "application/x-www-form-urlencoded"
	MIME_MultipartPOSTForm = "multipart/form-data"
)

type BindingInterface interface {
	Name() string
	Bind(*http.Request, interface{}) error
}

type Validator struct {
	*validator.Validate
}

func Default(method, contentType string) BindingInterface {
	if http.MethodGet == method {
		return Form
	}
	switch cleanContentType(contentType) {
	case MIME_JSON:
		return JSON
	case MIME_POSTForm:
		return FormPost
	case MIME_MultipartPOSTForm:
		return FormMultipart
	default:
		return Form
	}
}

func validate(obj interface{}) error {
	if DefaultValidator == nil {
		return nil
	}
	return DefaultValidator.Struct(obj)
}

func cleanContentType(contentType string) string {
	if i := strings.Index(contentType, ";"); i != -1 {
		return contentType[:i]
	}

	return contentType
}
