package binding

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Foo struct {
	Foo string `validate:"required" form:"foo"`
	Bar string `validate:"required" form:"bar"`
}

type TestMappingStruct struct {
	Int      int    `form:"int" default:"1"`
	Int8     int8   `form:"int8" default:"2"`
	Int16    int16  `form:"int16" default:"3"`
	Int32    int32  `form:"int32" default:"4"`
	Int64    int64  `form:"int64" default:"5"`
	Uint     uint   `form:"uint" default:"10"`
	Uint8    uint8  `form:"uint8" default:"11"`
	Uint16   uint16 `form:"uint16" default:"12"`
	Uint32   uint32 `form:"uint32" default:"13"`
	Uint64   uint64 `form:"uint64" default:"14"`
	String   string `form:"string" default:"hello world"`
	IntSlice []int  `form:"intSlice,split" default:"1,2,3"`
	Bool     bool   `form:"bool" default:"false"`
}

func TestBindingDefault(t *testing.T) {
	assert.Equal(t, Default("GET", ""), Form)
	assert.Equal(t, Default("GET", "application/json; charset=utf-8"), Form)
	assert.Equal(t, Default("GET", "application/x-www-form-urlencoded; charset=utf-8"), Form)
	assert.Equal(t, Default("GET", "multipart/form-data; charset=utf-8"), Form)

	assert.Equal(t, Default("POST", "application/json; charset=utf-8"), JSON)
	assert.Equal(t, Default("POST", "application/x-www-form-urlencoded; charset=utf-8"), FormPost)
	assert.Equal(t, Default("POST", "multipart/form-data; charset=utf-8"), FormMultipart)
}

func TestMapping(t *testing.T) {
	query := "/?int=12&int8=12&int16=12&int32=12&int64=12&uint=12&uint8=12&uint16=12&uint32=12&uint64=12&string=liu lang&intSlice=9,8,7,6,&bool=true"
	request1 := requestWithBody("GET", query, "", "int=24")
	request1.ParseForm()
	m := new(TestMappingStruct)
	mapForm(m, request1.Form)
	assert.Equal(t, m.Int, int(12))
	assert.Equal(t, m.Int8, int8(12))
	assert.Equal(t, m.Int16, int16(12))
	assert.Equal(t, m.Int32, int32(12))
	assert.Equal(t, m.Int64, int64(12))
	assert.Equal(t, m.Uint, uint(12))
	assert.Equal(t, m.Uint8, uint8(12))
	assert.Equal(t, m.Uint16, uint16(12))
	assert.Equal(t, m.Uint32, uint32(12))
	assert.Equal(t, m.Uint64, uint64(12))
	assert.Equal(t, m.String, "liu lang")
	assert.Equal(t, len(m.IntSlice), 4)
	assert.Equal(t, m.Bool, true)
	request2 := requestWithBody("GET", "/", "", "")
	request2.ParseForm()
	n := new(TestMappingStruct)
	mapForm(n, request2.Form)
	assert.Equal(t, n.Int, int(1))
	assert.Equal(t, n.Int8, int8(2))
	assert.Equal(t, n.Int16, int16(3))
	assert.Equal(t, n.Int32, int32(4))
	assert.Equal(t, n.Int64, int64(5))
	assert.Equal(t, n.Uint, uint(10))
	assert.Equal(t, n.Uint8, uint8(11))
	assert.Equal(t, n.Uint16, uint16(12))
	assert.Equal(t, n.Uint32, uint32(13))
	assert.Equal(t, n.Uint64, uint64(14))
	assert.Equal(t, n.String, "hello world")
	assert.Equal(t, len(n.IntSlice), 3)
	assert.Equal(t, n.Bool, false)
}

func TestCleanContentType(t *testing.T) {
	c1 := "application/json"
	c2 := "application/json; charset=utf-8"
	assert.Equal(t, cleanContentType(c1), c1)
	assert.Equal(t, cleanContentType(c2), "application/json")
}

func TestPOSTFormBinding(t *testing.T) {
	request := requestWithBody("POST", "/?foo=foo&bar=bar", MIME_POSTForm, "foo=hello&bar=world")
	foo := new(Foo)
	assert.Equal(t, FormPost.Bind(request, foo), nil)
	assert.Equal(t, foo.Foo, "hello")
	assert.Equal(t, foo.Bar, "world")
}

func TestMultipartFormBinding(t *testing.T) {
	request := multipartRequest()
	foo := new(Foo)
	assert.Equal(t, FormMultipart.Bind(request, foo), nil)
	assert.Equal(t, foo.Foo, "hello")
	assert.Equal(t, foo.Bar, "world")
}

func TestJSONBinding(t *testing.T) {
	request := requestWithBody("POST", "/?foo=foo&bar=bar", MIME_JSON, `{"foo":"hello","bar":"world"}`)
	foo := new(Foo)
	assert.Equal(t, JSON.Bind(request, foo), nil)
	assert.Equal(t, foo.Foo, "hello")
	assert.Equal(t, foo.Bar, "world")
}

func TestValidationFails(t *testing.T) {
	var obj Foo
	req := requestWithBody("POST", "/", MIME_JSON, `{"bar": "foo"}`)
	err := JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func BenchmarkBindingForm(b *testing.B) {
	req := requestWithBody("POST", "/", MIME_POSTForm, "foo=bar&bar=foo")
	req.Header.Add("Content-Type", MIME_POSTForm)
	f := Form
	for i := 0; i < b.N; i++ {
		obj := &Foo{}
		f.Bind(req, obj)
	}
}

func requestWithBody(method, url, contentType, body string) (req *http.Request) {
	req, _ = http.NewRequest(method, url, bytes.NewBufferString(body))
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	return
}

func multipartRequest() *http.Request {
	boundary := "--testboundary"
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	defer mw.Close()

	mw.SetBoundary(boundary)
	mw.WriteField("foo", "hello")
	mw.WriteField("bar", "world")
	req, _ := http.NewRequest("POST", "/?foo=foo&bar=bar", body)
	req.Header.Set("Content-Type", MIME_MultipartPOSTForm+"; boundary="+boundary)
	return req
}
