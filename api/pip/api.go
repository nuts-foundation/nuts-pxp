package pip

import (
	"context"
	"encoding/json"
	http2 "net/http"

	"github.com/nuts-foundation/nuts-pxp/db"
	"github.com/nuts-foundation/nuts-pxp/http"
)

var _ StrictServerInterface = (*Wrapper)(nil)
var _ http.Router = (*Wrapper)(nil)

type Wrapper struct {
	DB db.DB
}

func (w Wrapper) Routes(router *http2.ServeMux) {
	HandlerFromMux(NewStrictHandlerWithOptions(w, []StrictMiddlewareFunc{}, StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  http.ErrorHandlerFunc,
		ResponseErrorHandlerFunc: http.ErrorHandlerFunc,
	}), router)
}

func (w Wrapper) CreateData(_ context.Context, request CreateDataRequestObject) (CreateDataResponseObject, error) {
	// serialize authInput for storage
	authInput, _ := json.Marshal(request.Body.AuthInput)
	data := db.SQLData{
		Id:        request.Id,
		Client:    request.Body.ClientId,
		Scope:     request.Body.Scope,
		Verifier:  request.Body.VerifierId,
		AuthInput: string(authInput),
	}
	err := w.DB.Create(data)
	if err != nil {
		return nil, err
	}
	return CreateData204Response{}, nil
}

func (w Wrapper) GetData(ctx context.Context, request GetDataRequestObject) (GetDataResponseObject, error) {
	data, err := w.DB.Get(request.Id)
	if err != nil {
		return nil, err
	}
	// turn data into map[string]interface{}
	var authInput map[string]interface{}
	err = json.Unmarshal([]byte(data.AuthInput), &authInput)
	if err != nil {
		return nil, err
	}
	return GetData200JSONResponse{
		ClientId:   data.Client,
		Scope:      data.Scope,
		VerifierId: data.Verifier,
		AuthInput:  authInput,
	}, nil
}

func (w Wrapper) DeleteData(ctx context.Context, request DeleteDataRequestObject) (DeleteDataResponseObject, error) {
	err := w.DB.Delete(request.Id)
	if err != nil {
		return nil, err
	}
	return DeleteData204Response{}, nil
}
