// Package pip provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
package pip

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"
)

// Data defines model for Data.
type Data struct {
	// AuthInput Data used in OPA script
	AuthInput map[string]interface{} `json:"auth_input"`

	// ClientId client DID (for now)
	ClientId string `json:"client_id"`

	// Scope The scope. Corresponds to the auth scopes
	Scope string `json:"scope"`

	// VerifierId verifier DID (for now)
	VerifierId string `json:"verifier_id"`
}

// CreateDataJSONRequestBody defines body for CreateData for application/json ContentType.
type CreateDataJSONRequestBody = Data

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Delete data for given id
	// (DELETE /pip/{id})
	DeleteData(ctx echo.Context, id string) error
	// Get pip data for given ide
	// (GET /pip/{id})
	GetData(ctx echo.Context, id string) error
	// Add authorization data used for OPA evaluation
	// (POST /pip/{id})
	CreateData(ctx echo.Context, id string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// DeleteData converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteData(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	id = ctx.Param("id")

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.DeleteData(ctx, id)
	return err
}

// GetData converts echo context to params.
func (w *ServerInterfaceWrapper) GetData(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	id = ctx.Param("id")

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetData(ctx, id)
	return err
}

// CreateData converts echo context to params.
func (w *ServerInterfaceWrapper) CreateData(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	id = ctx.Param("id")

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.CreateData(ctx, id)
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

	router.DELETE(baseURL+"/pip/:id", wrapper.DeleteData)
	router.GET(baseURL+"/pip/:id", wrapper.GetData)
	router.POST(baseURL+"/pip/:id", wrapper.CreateData)

}

type DeleteDataRequestObject struct {
	Id string `json:"id"`
}

type DeleteDataResponseObject interface {
	VisitDeleteDataResponse(w http.ResponseWriter) error
}

type DeleteData204Response struct {
}

func (response DeleteData204Response) VisitDeleteDataResponse(w http.ResponseWriter) error {
	w.WriteHeader(204)
	return nil
}

type GetDataRequestObject struct {
	Id string `json:"id"`
}

type GetDataResponseObject interface {
	VisitGetDataResponse(w http.ResponseWriter) error
}

type GetData200JSONResponse Data

func (response GetData200JSONResponse) VisitGetDataResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type CreateDataRequestObject struct {
	Id   string `json:"id"`
	Body *CreateDataJSONRequestBody
}

type CreateDataResponseObject interface {
	VisitCreateDataResponse(w http.ResponseWriter) error
}

type CreateData204Response struct {
}

func (response CreateData204Response) VisitCreateDataResponse(w http.ResponseWriter) error {
	w.WriteHeader(204)
	return nil
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Delete data for given id
	// (DELETE /pip/{id})
	DeleteData(ctx context.Context, request DeleteDataRequestObject) (DeleteDataResponseObject, error)
	// Get pip data for given ide
	// (GET /pip/{id})
	GetData(ctx context.Context, request GetDataRequestObject) (GetDataResponseObject, error)
	// Add authorization data used for OPA evaluation
	// (POST /pip/{id})
	CreateData(ctx context.Context, request CreateDataRequestObject) (CreateDataResponseObject, error)
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

// DeleteData operation middleware
func (sh *strictHandler) DeleteData(ctx echo.Context, id string) error {
	var request DeleteDataRequestObject

	request.Id = id

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.DeleteData(ctx.Request().Context(), request.(DeleteDataRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "DeleteData")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(DeleteDataResponseObject); ok {
		return validResponse.VisitDeleteDataResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// GetData operation middleware
func (sh *strictHandler) GetData(ctx echo.Context, id string) error {
	var request GetDataRequestObject

	request.Id = id

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetData(ctx.Request().Context(), request.(GetDataRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetData")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(GetDataResponseObject); ok {
		return validResponse.VisitGetDataResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// CreateData operation middleware
func (sh *strictHandler) CreateData(ctx echo.Context, id string) error {
	var request CreateDataRequestObject

	request.Id = id

	var body CreateDataJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.CreateData(ctx.Request().Context(), request.(CreateDataRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "CreateData")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(CreateDataResponseObject); ok {
		return validResponse.VisitCreateDataResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}
