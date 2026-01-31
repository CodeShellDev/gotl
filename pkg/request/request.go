package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/codeshelldev/gotl/pkg/query"
)

const (
	Json    BodyType = "Json"
	Form    BodyType = "Form"
	Unknown BodyType = "Unknown"
)

type BodyType string

type Body struct {
	Data  map[string]any
	Raw   []byte
	Empty bool
}

// Body to string
func (body Body) ToString() string {
	return string(body.Raw)
}

// Write body into response writer
func (body *Body) Write(w http.ResponseWriter) error {
	newBody, err := CreateBody(body.Data)

	if err != nil {
		return err
	}

	*body = newBody

	bodyLength := len(body.Raw)

	w.Header().Set("Content-Length", strconv.Itoa(bodyLength))

	if !body.Empty {
		w.Header().Set("Content-Type", "application/json")
	}

	w.Write(body.Raw)

	return nil
}

// Update body in request
func (body *Body) UpdateReq(req *http.Request) error {
	newBody, err := CreateBody(body.Data)

	if err != nil {
		return err
	}

	*body = newBody

	bodyLength := len(body.Raw)

	req.ContentLength = int64(bodyLength)
	req.Header.Set("Content-Length", strconv.Itoa(bodyLength))

	if !body.Empty {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Body = io.NopCloser(bytes.NewReader(body.Raw))

	return nil
}

// Update body in response
func (body *Body) UpdateRes(res *http.Response) error {
	newBody, err := CreateBody(body.Data)

	if err != nil {
		return err
	}

	*body = newBody

	bodyLength := len(body.Raw)

	res.ContentLength = int64(bodyLength)
	res.Header.Set("Content-Length", strconv.Itoa(bodyLength))

	if !body.Empty {
		res.Header.Set("Content-Type", "application/json")
	}

	res.Body = io.NopCloser(bytes.NewReader(body.Raw))

	return nil
}

// Create new body with data
func CreateBody(data map[string]any) (Body, error) {
	if len(data) <= 0 {
		err := errors.New("empty data map")

		return Body{Empty: true}, err
	}

	bytes, err := json.Marshal(data)

	if err != nil {

		return Body{Empty: true}, err
	}

	isEmpty := len(data) <= 0

	return Body{
		Data:  data,
		Raw:   bytes,
		Empty: isEmpty,
	}, nil
}

// Read body from request
func ReadReqBody(req *http.Request) ([]byte, error) {
	bodyBytes, err := readBodyBytes(req.Body)

	req.Body.Close()

	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

// Read body from response
func ReadResBody(res *http.Response) ([]byte, error) {
	bodyBytes, err := readBodyBytes(res.Body)

	res.Body.Close()

	res.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func readBodyBytes(body io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(body, 5<<20))
}

// Get headers from request
func GetReqHeaders(req *http.Request) map[string][]string {
	data := map[string][]string{}

	maps.Copy(data, req.Header)

	return data
}

// Get headers from response
func GetResHeaders(res *http.Response) map[string][]string {
	data := map[string][]string{}

	maps.Copy(data, res.Header)

	return data
}

// Parse headers into `map[string]any`
func ParseHeaders(headers map[string][]string) map[string]any {
	generic := make(map[string]any, len(headers))

	for i, header := range headers {
		if len(header) == 1 {
			generic[i] = header[0]
		} else {
			generic[i] = header
		}
	}

	return generic
}

// Outputs a URL object with fields populated from the request
func ParseReqURL(req *http.Request) (*url.URL, error) {
	scheme := "http"

	if req.TLS != nil {
		scheme = "https"
	}

	return url.Parse(scheme + "://" + req.Host + req.URL.RequestURI())
}

// Get body from request
func GetReqBody(req *http.Request) (Body, error) {
	bytes, err := ReadReqBody(req)

	var isEmpty bool

	if err != nil {
		return Body{Empty: true}, err
	}

	if len(bytes) <= 0 {
		return Body{Empty: true}, nil
	}

	var data map[string]any

	contentType := req.Header.Get("Content-Type")

	switch getBodyType(contentType) {
	case Json:
		data, err = getJsonData(bytes)

		if err != nil {
			return Body{Empty: true}, err
		}
	case Form:
		data, err = getFormData(bytes)

		if err != nil {
			return Body{Empty: true}, err
		}
	}

	isEmpty = len(data) <= 0

	return Body{
		Raw:   bytes,
		Data:  data,
		Empty: isEmpty,
	}, nil
}

// Get body from response
func GetResBody(res *http.Response) (Body, error) {
	bytes, err := ReadResBody(res)

	var isEmpty bool

	if err != nil {
		return Body{Empty: true}, err
	}

	if len(bytes) <= 0 {
		return Body{Empty: true}, nil
	}

	var data map[string]any

	contentType := res.Header.Get("Content-Type")

	switch getBodyType(contentType) {
	case Json:
		data, err = getJsonData(bytes)

		if err != nil {
			return Body{Empty: true}, err
		}
	case Form:
		data, err = getFormData(bytes)

		if err != nil {
			return Body{Empty: true}, err
		}
	}

	isEmpty = len(data) <= 0

	return Body{
		Raw:   bytes,
		Data:  data,
		Empty: isEmpty,
	}, nil
}


func getBodyType(contentType string) BodyType {
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		return Json

	case strings.HasPrefix(contentType, "multipart/form-data"):
		return Form

	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		return Form
	default:
		return Unknown
	}
}

func getJsonData(body []byte) (map[string]any, error) {
	var data map[string]any

	err := json.Unmarshal(body, &data)

	if err != nil {

		return nil, err
	}

	return data, nil
}

func getFormData(body []byte) (map[string]any, error) {
	return query.ParseTypedQuery(string(body))
}