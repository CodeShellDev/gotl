package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
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

// Write body into request
func (body *Body) Write(req *http.Request) error {
	rawBytes, err := json.Marshal(body.Data)
	if err != nil {
		return err
	}
	
	body.Raw = rawBytes

	bodyLength := len(body.Raw)

	req.ContentLength = int64(bodyLength)
	req.Header.Set("Content-Length", strconv.Itoa(bodyLength))

	req.Body = io.NopCloser(bytes.NewReader(body.Raw))

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
func ReadBody(req *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(io.LimitReader(req.Body, 5<<20))

	req.Body.Close()

	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

// Get headers from request
func GetReqHeaders(req *http.Request) map[string][]string {
	data := map[string][]string{}

	maps.Copy(data, req.Header)

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

// Get body from request
func GetReqBody(req *http.Request) (Body, error) {
	bytes, err := ReadBody(req)

	var isEmpty bool

	if err != nil {
		return Body{Empty: true}, err
	}

	if len(bytes) <= 0 {
		return Body{Empty: true}, nil
	}

	var data map[string]any

	switch getBodyType(req) {
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

func getBodyType(req *http.Request) BodyType {
	contentType := req.Header.Get("Content-Type")

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