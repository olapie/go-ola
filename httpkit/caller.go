package httpkit

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"go.olapie.com/ola/mimetypes"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	"go.olapie.com/ola/headers"
	"go.olapie.com/security/base62"

	"go.olapie.com/ola/errorutil"

	"go.olapie.com/ola/urlutil"
	"go.olapie.com/utils"
)

type void = struct{}

type CallResult[R any] struct {
	Value  R
	Header http.Header
	Error  error
}

type Caller[IN any, OUT any] struct {
	Client     *http.Client
	Method     string
	Endpoint   string
	BeforeCall func(req *http.Request) error
}

func NewCaller[IN any, OUT any](method string, endpoint string) *Caller[IN, OUT] {
	e := &Caller[IN, OUT]{
		Method:   method,
		Endpoint: endpoint,
	}
	return e
}

func (c *Caller[IN, OUT]) WithQuery(query url.Values) *Caller[IN, OUT] {
	cc := *c
	var err error
	cc.Endpoint, err = urlutil.AppendQuery(c.Endpoint, query)
	if err != nil {
		log.Println("http.Caller.WithQuery", err)
	}
	return &cc
}

func (c *Caller[IN, OUT]) WithQueryArgs(keysAndValues ...any) *Caller[IN, OUT] {
	n := len(keysAndValues)
	if n%2 != 0 {
		panic("keyAndValues is not paired")
	}

	query := url.Values{}
	for i := 0; i < n; i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		ks, ok := k.(string)
		if !ok {
			if stringer, ok := k.(fmt.Stringer); ok {
				ks = stringer.String()
			}
		}

		if ks == "" {
			panic(fmt.Sprintf("keysAndValues[%d] is not a string key", i))
		}

		vs, ok := v.(string)
		if !ok {
			if stringer, ok := v.(fmt.Stringer); ok {
				vs = stringer.String()
			} else if utils.IsNumber(v) {
				vs = fmt.Sprint(v)
			}
		}
		if vs == "" {
			panic(fmt.Sprintf("keysAndValues[%d] is not a string or number value", i+1))
		}
		query.Set(ks, vs)
	}
	return c.WithQuery(query)
}

func (c *Caller[IN, OUT]) Call(ctx context.Context, input IN) (OUT, error) {
	var out OUT
	resp, err := c.call(ctx, input)
	if err != nil {
		return out, err
	}
	return GetResponseResult[OUT](resp)
}

func (c *Caller[IN, OUT]) GetResult(ctx context.Context, input IN) *CallResult[OUT] {
	res := new(CallResult[OUT])
	resp, err := c.call(ctx, input)
	if err != nil {
		res.Error = err
		return res
	}
	res.Header = resp.Header
	out, err := GetResponseResult[OUT](resp)
	if err != nil {
		res.Error = err
		return res
	}
	res.Value = out
	return res
}

func (c *Caller[IN, OUT]) CallAndRewrite(ctx context.Context, input IN, writer io.Writer) error {
	resp, err := c.call(ctx, input)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, resp.Body)
	return err
}

func (c *Caller[IN, OUT]) call(ctx context.Context, input IN) (*http.Response, error) {
	var contentType string
	endpoint, err := url.PathUnescape(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("unescape path: %w", err)
	}
	body, err := c.parseInput(&contentType, &endpoint, input)
	if err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, c.Method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request %s %s: %w", c.Method, endpoint, err)
	}
	headers.SetContentType(req.Header, contentType)
	headers.SetTraceID(req.Header, base62.NewUUIDString())

	client := http.DefaultClient
	if c.Client != nil {
		client = c.Client
	}

	if c.BeforeCall != nil {
		if err = c.BeforeCall(req); err != nil {
			return nil, fmt.Errorf("before call: %w", err)
		}
	}

	log.Println(req.Method, req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		if err == context.DeadlineExceeded {
			err = errorutil.NewError(http.StatusRequestTimeout, err.Error())
		} else {
			if tr, ok := err.(interface{ Timeout() bool }); ok && tr.Timeout() {
				err = errorutil.NewError(http.StatusRequestTimeout, err.Error())
			}
		}
		return nil, fmt.Errorf("send request: %w", err)
	}
	return resp, nil
}

func (c *Caller[IN, OUT]) parseInput(contentType *string, endpoint *string, input any) (io.Reader, error) {
	if input == nil {
		return nil, nil
	}

	if b, ok := input.([]byte); ok {
		return bytes.NewReader(b), nil
	}

	body, ok := input.(io.Reader)
	if ok {
		if *contentType == "" {
			*contentType = mimetypes.OctetStream
		}
		return body, nil
	}

	if v, ok := input.(url.Values); ok {
		newEndpoint, err := urlutil.AppendQuery(*endpoint, v)
		if err != nil {
			return nil, err
		}
		*endpoint = newEndpoint
		return nil, nil
	}

	newEndpoint, remain := urlutil.SetPathParams(*endpoint, input)
	*endpoint = newEndpoint

	if remain == nil {
		return nil, nil
	}

	kindOfRemain := utils.IndirectKind(remain)
	switch kindOfRemain {
	case reflect.Struct, reflect.Map, reflect.Slice:
		*contentType = mimetypes.JsonUTF8
		data, err := json.Marshal(input)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
		return bytes.NewBuffer(data), nil
	default:
		if utils.IsNumber(remain) || utils.IsString(remain) {
			*contentType = mimetypes.PlainUTF8
			return bytes.NewReader([]byte(fmt.Sprint(remain))), nil
		}
		return nil, fmt.Errorf("unsupported value type: %T", input)
	}
}

func NewGet[IN any, OUT any](endpoint string) *Caller[IN, OUT] {
	return NewCaller[IN, OUT](http.MethodGet, endpoint)
}

func NewPost[IN any, OUT any](endpoint string) *Caller[IN, OUT] {
	return NewCaller[IN, OUT](http.MethodPost, endpoint)
}

func NewPut[IN any, OUT any](endpoint string) *Caller[IN, OUT] {
	return NewCaller[IN, OUT](http.MethodPut, endpoint)
}

func NewPatch[IN any, OUT any](endpoint string) *Caller[IN, OUT] {
	return NewCaller[IN, OUT](http.MethodPatch, endpoint)
}

func NewDelete[IN any](endpoint string) *Caller[IN, void] {
	return NewCaller[IN, void](http.MethodDelete, endpoint)
}

func NewHead(endpoint string) *Caller[void, void] {
	return NewCaller[void, void](http.MethodHead, endpoint)
}

func NewOptions(endpoint string) *Caller[void, void] {
	return NewCaller[void, void](http.MethodOptions, endpoint)
}

func NewTrace[T any](endpoint string) *Caller[T, T] {
	return NewCaller[T, T](http.MethodTrace, endpoint)
}

func NewConnect(endpoint string) *Caller[void, void] {
	return NewCaller[void, void](http.MethodConnect, endpoint)
}

type UnmarshalFunc func([]byte, any) error

var contentTypeToUnmarshalFunc sync.Map

func init() {
	RegisterUnmarshalFunc(mimetypes.JSON, json.Unmarshal)
	RegisterUnmarshalFunc(mimetypes.JsonUTF8, json.Unmarshal)
	RegisterUnmarshalFunc(mimetypes.XML, xml.Unmarshal)
	RegisterUnmarshalFunc(mimetypes.XML2, xml.Unmarshal)
	RegisterUnmarshalFunc(mimetypes.XmlUTF8, xml.Unmarshal)
}

func RegisterUnmarshalFunc(contentType string, f UnmarshalFunc) {
	contentTypeToUnmarshalFunc.Store(contentType, f)
}

func GetUnmarshalFunc(contentType string) UnmarshalFunc {
	v, ok := contentTypeToUnmarshalFunc.Load(contentType)
	if ok {
		u, _ := v.(UnmarshalFunc)
		return u
	}
	return nil
}

func GetResponseResult[T any](resp *http.Response) (T, error) {
	var res T
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return res, fmt.Errorf("read resp body: %v", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return res, errorutil.NewError(resp.StatusCode, string(body))
	}

	if any(res) == nil {
		return res, nil
	}

	if val := reflect.ValueOf(res); val.Kind() == reflect.Struct && val.Type().NumField() == 0 {
		return res, nil
	}

	ct := headers.GetContentType(resp.Header)
	if f := GetUnmarshalFunc(ct); f != nil {
		err = f(body, &res)
		return res, err
	}

	if len(body) == 0 {
		err = errors.New("no data")
	} else if _, ok := any(res).([]byte); ok {
		res = any(body).(T)
	} else {
		if err = utils.SetBytes(&res, body); err != nil {
			err = fmt.Errorf("cannot handle %s: %w ", ct, err)
		}
	}
	return res, err
}
