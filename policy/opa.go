package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nuts-foundation/nuts-pxp/config"
	"github.com/nuts-foundation/nuts-pxp/db"
	"github.com/open-policy-agent/opa/rego"
)

var _ DecisionMaker = (*OPADecision)(nil)

type OPADecision struct {
	db            db.DB
	modules       map[string]string
	scopeToModule map[string]string
}

func New(config config.Config, db db.DB) (*OPADecision, error) {
	decisionMaker := &OPADecision{
		db:            db,
		modules:       make(map[string]string),
		scopeToModule: make(map[string]string),
	}

	// load rego files from the policy directory
	if err := decisionMaker.load(config); err != nil {
		return nil, err
	}
	return decisionMaker, nil
}

func (dm *OPADecision) Query(ctx context.Context, requestLine map[string]interface{}, introspectionResult map[string]interface{}) (bool, error) {
	// extract scope, client and verifier from introspectionResult
	scope, ok := introspectionResult["scope"].(string)
	if !ok {
		return false, fmt.Errorf("no scope in introspection result")
	}
	client, ok := introspectionResult["client_id"].(string)
	if !ok {
		return false, fmt.Errorf("no client_id in introspection result")
	}
	verifier, ok := introspectionResult["sub"].(string)
	if !ok {
		return false, fmt.Errorf("no sub in introspection result")
	}

	// query DB for runtime data
	data, err := dm.db.Query(scope, verifier, client)
	if err != nil {
		return false, err
	}

	// parse the data into a map
	input := map[string]interface{}{}
	external := map[string]interface{}{}
	request := map[string]interface{}{}
	err = json.Unmarshal([]byte(data), &external)
	if err != nil {
		return false, err
	}
	for k, v := range requestLine {
		request[k] = v
	}
	// merge the request line and introspection result into the input
	input["external"] = external
	input["request"] = request
	for k, v := range introspectionResult {
		input[k] = v
	}
	return dm.query(ctx, scope, input)
}

// Query as defined on https://www.openpolicyagent.org/docs/latest/integration/#integrating-with-the-go-api
func (dm *OPADecision) query(ctx context.Context, scope string, input map[string]interface{}) (bool, error) {
	// make a decision using the OPA engine
	// we assume scope == module == file name
	// package is parsed when loading the file and mapped to the scope
	query, err := rego.New(
		rego.Query(fmt.Sprintf("x = data.%s.allow", dm.scopeToModule[scope])),
		rego.Module(scope, dm.modules[scope]),
	).PrepareForEval(ctx)

	if err != nil {
		return false, err
	}
	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, nil
	} else if len(results) == 0 {
		return false, nil
	} else if result, ok := results[0].Bindings["x"].(bool); !ok {
		return false, nil
	} else {
		return result, nil
	}
}

func (dm *OPADecision) load(config config.Config) error {
	// Load policy files from the directory specified in the configuration.
	// loop over all rego files located in config.PolicyDir
	// and load them into the OPA engine
	// return an error if the policy files cannot be loaded
	dir, err := os.Open(config.PolicyDir)
	if err != nil {
		return err
	}
	defer dir.Close()

	// read all the files in the directory
	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	// load all the files
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".rego") {
			continue
		}
		// split filename, remove .rego extension
		module := strings.TrimSuffix(file.Name(), ".rego")
		err := dm.loadFromFile(fmt.Sprintf("%s/%s", config.PolicyDir, file.Name()), module)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dm *OPADecision) loadFromFile(file string, module string) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	asString := string(bytes)
	// scan rego for package name
	// first split the rego file into lines
	lines := strings.Split(asString, "\n")
	// then find the package line
	for _, line := range lines {
		if strings.HasPrefix(line, "package") {
			// extract the package name
			packageName := strings.TrimSpace(strings.TrimPrefix(line, "package"))
			// store the package name
			dm.scopeToModule[module] = packageName
			break
		}
	}

	dm.modules[module] = string(bytes)
	return nil
}
