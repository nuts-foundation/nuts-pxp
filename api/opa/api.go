package opa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-pxp/policy"
)

var _ StrictServerInterface = (*Wrapper)(nil)

type Wrapper struct {
	DecisionMaker policy.DecisionMaker
}

func (w Wrapper) EvaluateDocumentWildcardPolicy(ctx context.Context, request EvaluateDocumentWildcardPolicyRequestObject) (EvaluateDocumentWildcardPolicyResponseObject, error) {
	if request.Body == nil {
		return nil, errors.New("missing body")
	}
	// APISIX combines the 'openid-connect' and 'opa' plugin results into the following body:
	//{
	//	"input": {
	//		"var": {
	//			"server_port": "9080",
	//			"remote_addr": "172.90.10.2",
	//			"timestamp": 1718289289,
	//			"remote_port": "54228",
	//			"server_addr": "172.90.10.12"
	//		},
	//		"type": "http",
	//		"request": {
	//			"scheme": "http",
	//			"method": "POST",
	//			"host": "pep-right",
	//			"query": {},
	//			"path": "/web/external/transfer/notify/21189b43-04d5-4f4f-86ed-e5ae21a87f84",
	//			"headers": {
	//				"X-Userinfo": "eyJvcmdhbml6YXRpb25fbmFtZSI6IkxlZnQiLCJzY29wZSI6ImVPdmVyZHJhY2h0LXJlY2VpdmVyIiwic3ViIjoiZGlkOndlYjpub2RlLnJpZ2h0LmxvY2FsOmlhbTpyaWdodCIsImV4cCI6MTcxODI5MDE4NiwiaWF0IjoxNzE4Mjg5Mjg2LCJpc3MiOiJkaWQ6d2ViOm5vZGUucmlnaHQubG9jYWw6aWFtOnJpZ2h0IiwiYWN0aXZlIjp0cnVlLCJjbGllbnRfaWQiOiJkaWQ6d2ViOm5vZGUubGVmdC5sb2NhbDppYW06bGVmdCIsIm9yZ2FuaXphdGlvbl9jaXR5IjoiR3JvZW5sbyJ9",
	//				"host": "pep-right:9080",
	//				"authorization": "Bearer TonUNXLwVn2UgJgVfpVDNa7WaXAlE2W-mS6CfqDzeP0",
	//				"content-length": "0",
	//				"user-agent": "go-resty/2.13.1 (https://github.com/go-resty/resty)",
	//				"X-Access-Token": "TonUNXLwVn2UgJgVfpVDNa7WaXAlE2W-mS6CfqDzeP0",
	//				"accept-encoding": "gzip",
	//				"content-type": "text/plain; charset=utf-8",
	//				"connection": "close"
	//			},
	//			"port": 9080
	//		}
	//	}
	//}
	outcome, err := w.handleEvaluate(ctx, *request.Body)
	if err != nil {
		return nil, err
	}

	// Expected response by APISIX is of the form:
	//{
	//	"result": {
	//		"allow": true
	//	}
	//}
	return EvaluateDocumentWildcardPolicy200JSONResponse(*outcome), nil
}

func (w Wrapper) EvaluateDocument(ctx context.Context, request EvaluateDocumentRequestObject) (EvaluateDocumentResponseObject, error) {
	if request.Body == nil {
		return nil, errors.New("missing body")
	}
	//fmt.Printf("%v\n", *request.Body)

	outcome, err := w.handleEvaluate(ctx, *request.Body)
	if err != nil {
		return nil, err
	}

	return EvaluateDocument200JSONResponse(*outcome), nil
}

func (w Wrapper) handleEvaluate(ctx context.Context, input Input) (*Outcome, error) {

	httpRequest, ok := input.Input["request"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid request, missing 'input.request'")
	}

	httpHeaders, ok := httpRequest["headers"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid request, missing 'input.request.headers'")
	}
	xUserinfoBase64, ok := httpHeaders["X-Userinfo"].(string)
	if !ok {
		return nil, errors.New("invalid request, missing 'input.request.headers.X-Userinfo is not a string'")
	}
	xUserinfoJSON, err := base64.URLEncoding.DecodeString(xUserinfoBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid request, failed to base64 decode X-Userinfo: %w", err)
	}
	xUserinfo := map[string]interface{}{}
	err = json.Unmarshal(xUserinfoJSON, &xUserinfo)
	if err != nil {
		return nil, fmt.Errorf("invalid request, failed to unmarshal X-Userinfo: %w", err)
	}

	decision, err := w.DecisionMaker.Query(ctx, httpRequest, xUserinfo)
	if err != nil {
		return nil, err
	}
	return &Outcome{Result: map[string]interface{}{"allow": decision}}, nil
}
