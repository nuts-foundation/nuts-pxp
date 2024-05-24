package policy

import "context"

// DecisionMaker is an interface that can be implemented by any policy decision maker.
type DecisionMaker interface {
	Query(ctx context.Context, requestLine map[string]interface{}, introspectionResult map[string]interface{}) (bool, error)
}
