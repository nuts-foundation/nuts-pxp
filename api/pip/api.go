package pip

import "context"

var _ StrictServerInterface = (*Wrapper)(nil)

type Wrapper struct{}

func (w Wrapper) CreateData(ctx context.Context, request CreateDataRequestObject) (CreateDataResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
