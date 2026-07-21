package nodes

import (
	"context"
	"encoding/json"

	"christiangeorgelucas/json-logic-tools/axiom"
	gen "christiangeorgelucas/json-logic-tools/gen"
	jsonlogic "github.com/diegoholiveira/jsonlogic/v3"
)

// ResolveVariables performs a partial evaluation: it returns the rule with
// every {"var": ...} reference substituted by its value from data, while
// otherwise preserving the JSON Logic structure (operators are NOT
// evaluated) — useful for auditing or debugging what a rule will actually
// see before running Evaluate. Wraps the library's own
// GetJsonLogicWithSolvedVars. Same bounds and error contract as Evaluate.
func ResolveVariables(ctx context.Context, ax axiom.Context, input *gen.JsonLogicRule) (*gen.JsonLogicResult, error) {
	if input.Logic == "" {
		return &gen.JsonLogicResult{Error: "logic is required"}, nil
	}
	if err := checkJSONBounds("logic", []byte(input.Logic), maxLogicBytes); err != nil {
		return &gen.JsonLogicResult{Error: err.Error()}, nil
	}

	dataRaw := input.Data
	if dataRaw == "" {
		dataRaw = "{}"
	}
	if err := checkJSONBounds("data", []byte(dataRaw), maxDataBytes); err != nil {
		return &gen.JsonLogicResult{Error: err.Error()}, nil
	}

	resolved, err := jsonlogic.GetJsonLogicWithSolvedVars(json.RawMessage(input.Logic), json.RawMessage(dataRaw))
	if err != nil {
		return &gen.JsonLogicResult{Error: err.Error()}, nil
	}

	return &gen.JsonLogicResult{Result: string(resolved)}, nil
}
