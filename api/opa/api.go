package opa

import "context"

var _ StrictServerInterface = (*Wrapper)(nil)

type Wrapper struct{}

func (w Wrapper) EvaluateDocument(ctx context.Context, request EvaluateDocumentRequestObject) (EvaluateDocumentResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
