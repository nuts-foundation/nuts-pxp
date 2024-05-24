package opa

import (
	"context"
	"fmt"
	"github.com/nuts-foundation/nuts-pxp/policy"
	"strings"
)

var _ StrictServerInterface = (*Wrapper)(nil)

type Wrapper struct {
	DecisionMaker policy.DecisionMaker
}

func (w Wrapper) EvaluateDocument(ctx context.Context, request EvaluateDocumentRequestObject) (EvaluateDocumentResponseObject, error) {
	// parse the requestLine and extract the method and path
	// the requestLine is formatted as an HTTP request line
	// e.g. "GET /api/v1/resource HTTP/1.1"
	// we are only interested in the method and path
	method, path, err := parseRequestLine(request.Params.Request)
	if err != nil {
		return nil, err
	}
	httpRequest := map[string]interface{}{}
	httpRequest["method"] = method
	httpRequest["path"] = path

	descision, err := w.DecisionMaker.Query(ctx, httpRequest, request.Params.XUserinfo)
	if err != nil {
		return nil, err
	}
	return EvaluateDocument200JSONResponse{Allow: descision}, nil
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
