// Package opa provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
package opa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"
)

// Input defines model for Input.
type Input struct {
	// Input Policy decision information. Must contain the fields in the example.
	Input map[string]interface{} `json:"input"`
}

// Outcome defines model for Outcome.
type Outcome struct {
	// Result The result of the OPA policy evaluation
	Result map[string]interface{} `json:"result"`
}

// EvaluateDocumentJSONRequestBody defines body for EvaluateDocument for application/json ContentType.
type EvaluateDocumentJSONRequestBody = Input

// EvaluateDocumentWildcardPolicyJSONRequestBody defines body for EvaluateDocumentWildcardPolicy for application/json ContentType.
type EvaluateDocumentWildcardPolicyJSONRequestBody = Input

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// calls https://www.openpolicyagent.org/docs/latest/rest-api/#get-a-document-with-input internally
	// (POST /v1/data)
	EvaluateDocument(ctx echo.Context) error
	// calls https://www.openpolicyagent.org/docs/latest/rest-api/#get-a-document-with-input internally
	// (POST /v1/data/*)
	EvaluateDocumentWildcardPolicy(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// EvaluateDocument converts echo context to params.
func (w *ServerInterfaceWrapper) EvaluateDocument(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.EvaluateDocument(ctx)
	return err
}

// EvaluateDocumentWildcardPolicy converts echo context to params.
func (w *ServerInterfaceWrapper) EvaluateDocumentWildcardPolicy(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.EvaluateDocumentWildcardPolicy(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.POST(baseURL+"/v1/data", wrapper.EvaluateDocument)
	router.POST(baseURL+"/v1/data/*", wrapper.EvaluateDocumentWildcardPolicy)

}

type EvaluateDocumentRequestObject struct {
	Body *EvaluateDocumentJSONRequestBody
}

type EvaluateDocumentResponseObject interface {
	VisitEvaluateDocumentResponse(w http.ResponseWriter) error
}

type EvaluateDocument200JSONResponse Outcome

func (response EvaluateDocument200JSONResponse) VisitEvaluateDocumentResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type EvaluateDocumentWildcardPolicyRequestObject struct {
	Body *EvaluateDocumentWildcardPolicyJSONRequestBody
}

type EvaluateDocumentWildcardPolicyResponseObject interface {
	VisitEvaluateDocumentWildcardPolicyResponse(w http.ResponseWriter) error
}

type EvaluateDocumentWildcardPolicy200JSONResponse Outcome

func (response EvaluateDocumentWildcardPolicy200JSONResponse) VisitEvaluateDocumentWildcardPolicyResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// calls https://www.openpolicyagent.org/docs/latest/rest-api/#get-a-document-with-input internally
	// (POST /v1/data)
	EvaluateDocument(ctx context.Context, request EvaluateDocumentRequestObject) (EvaluateDocumentResponseObject, error)
	// calls https://www.openpolicyagent.org/docs/latest/rest-api/#get-a-document-with-input internally
	// (POST /v1/data/*)
	EvaluateDocumentWildcardPolicy(ctx context.Context, request EvaluateDocumentWildcardPolicyRequestObject) (EvaluateDocumentWildcardPolicyResponseObject, error)
}

type StrictHandlerFunc = strictecho.StrictEchoHandlerFunc
type StrictMiddlewareFunc = strictecho.StrictEchoMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// EvaluateDocument operation middleware
func (sh *strictHandler) EvaluateDocument(ctx echo.Context) error {
	var request EvaluateDocumentRequestObject

	var body EvaluateDocumentJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.EvaluateDocument(ctx.Request().Context(), request.(EvaluateDocumentRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "EvaluateDocument")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(EvaluateDocumentResponseObject); ok {
		return validResponse.VisitEvaluateDocumentResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// EvaluateDocumentWildcardPolicy operation middleware
func (sh *strictHandler) EvaluateDocumentWildcardPolicy(ctx echo.Context) error {
	var request EvaluateDocumentWildcardPolicyRequestObject

	var body EvaluateDocumentWildcardPolicyJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.EvaluateDocumentWildcardPolicy(ctx.Request().Context(), request.(EvaluateDocumentWildcardPolicyRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "EvaluateDocumentWildcardPolicy")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(EvaluateDocumentWildcardPolicyResponseObject); ok {
		return validResponse.VisitEvaluateDocumentWildcardPolicyResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}
