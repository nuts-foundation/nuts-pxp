package opa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	http2 "net/http"
	"net/http/httputil"
	"strings"

	"github.com/nuts-foundation/nuts-pxp/http"
	"github.com/nuts-foundation/nuts-pxp/policy"
)

var _ StrictServerInterface = (*Wrapper)(nil)
var _ http.Router = (*Wrapper)(nil)

type Wrapper struct {
	DecisionMaker policy.DecisionMaker
}

func (w Wrapper) Routes(router *http2.ServeMux) {
	handler := NewStrictHandlerWithOptions(w, []StrictMiddlewareFunc{}, StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  http.ErrorHandlerFunc,
		ResponseErrorHandlerFunc: http.ErrorHandlerFunc,
	})
	HandlerFromMux(handler, router)
	router.HandleFunc("POST /v1/data/*", func(writer http2.ResponseWriter, request *http2.Request) {
		req, _ := httputil.DumpRequest(request, true)
		fmt.Printf("REQUEST DUMP START: %s\n", req)
		fmt.Println("REQUEST DUMP END")
		// we validate the path and if valid we route it to /v1/data
		if err := validatePath(request.URL.Path); err != nil {
			http.ErrorHandlerFunc(writer, request, err)
			// MUST return here if the path is invalid
			return
		}
		request.URL.Path = "/v1/data"

		wrapper := ServerInterfaceWrapper{
			Handler:            handler,
			HandlerMiddlewares: []MiddlewareFunc{},
			ErrorHandlerFunc:   http.ErrorHandlerFunc,
		}
		wrapper.EvaluateDocument(writer, request)
	})
}

func validatePath(path string) error {
	// path matches /v1/data/*, we validate *
	// /v1/data/{package}/{decision}
	// OPA 'package' contains 1 or more path elements.
	//		If there is more than 1 element, replace '/' with '.' for the package name. This value is currently not validated.
	// 'decision' is the variable name of the boolean in the result containing the OPA policy decision.
	// 		If * contains more than 1 path element we assume the last element is the decision value, which MUST be equal to 'allow'
	parts := strings.Split(path, "/") // ["", "v1", "data", ...]
	if len(parts) > 4 && parts[len(parts)-1] != "allow" {
		return errors.New("invalid OPA request")
	}
	return nil
}

func (w Wrapper) EvaluateDocument(ctx context.Context, request EvaluateDocumentRequestObject) (EvaluateDocumentResponseObject, error) {
	// parse the requestLine and extract the method and path
	// the requestLine is formatted as an HTTP request line
	// e.g. "GET /api/v1/resource HTTP/1.1"
	// we are only interested in the method and path
	//method, path, err := parseRequestLine(request.Params.Request)
	//if err != nil {
	//	return nil, err
	//}
	//httpRequest := map[string]interface{}{}
	//httpRequest["method"] = method
	//httpRequest["path"] = path

	// request.Body =:
	// {
	//		"input": {
	//			"request": {
	//				"method": ...,
	//				"path": ...,
	//			}, {
	//			"headers": {
	//				"X-Userinfo": ...,
	//			},
	//	 	},
	// }
	httpRequest := (*request.Body)["input"].(map[string]interface{})["request"].(map[string]interface{})
	httpHeaders := httpRequest["headers"].(map[string]interface{})
	xUserinfoBase64 := httpHeaders["X-Userinfo"].(string)
	xUserinfoJSON, err := base64.URLEncoding.DecodeString(xUserinfoBase64)
	if err != nil {
		panic(err)
	}
	xUserinfo := map[string]interface{}{}
	err = json.Unmarshal(xUserinfoJSON, &xUserinfo)
	if err != nil {
		panic(err)
	}

	descision, err := w.DecisionMaker.Query(ctx, httpRequest, xUserinfo)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{"allow": descision}
	return EvaluateDocument200JSONResponse{Result: result}, nil
}

// parseRequestLine parses the request line and extracts the method and path
// e.g. "GET /api/v1/resource HTTP/1.1" -> "GET", "/api/v1/resource"
func parseRequestLine(requestLine string) (method, path string, err error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid request line: %s", requestLine)
	}
	return parts[0], parts[1], nil
}
