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
// GetJsonLogicWithSolvedVars, which requires an object (map) at the rule's
// root — a bare literal/array root (e.g. "true", "[1,2]") is rejected with
// a structured error rather than attempted, since the library itself has
// no fallback for that shape. Same bounds and error contract as Evaluate,
// including the same evaluation time limit.
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

	// GetJsonLogicWithSolvedVars does an unchecked `rule.(map[string]any)`
	// type assertion on the root value internally, with no panic recovery
	// of its own — a bare literal or array root would otherwise crash the
	// whole process, not just this request. Reject that shape up front.
	var parsedLogic any
	if err := json.Unmarshal([]byte(input.Logic), &parsedLogic); err != nil {
		return &gen.JsonLogicResult{Error: "logic is not valid JSON: " + err.Error()}, nil
	}
	if _, ok := parsedLogic.(map[string]any); !ok {
		return &gen.JsonLogicResult{Error: "ResolveVariables requires an object (\"{...}\") at the rule's root; a bare literal or array root has no variables to resolve — use Evaluate instead"}, nil
	}

	out, err := runWithTimeoutAndRecover(evalTimeout, func() (any, error) {
		resolved, err := jsonlogic.GetJsonLogicWithSolvedVars(json.RawMessage(input.Logic), json.RawMessage(dataRaw))
		return resolved, err
	})
	if err != nil {
		return &gen.JsonLogicResult{Error: err.Error()}, nil
	}

	// GetJsonLogicWithSolvedVars returns []byte (json.RawMessage is a
	// distinct named type, so assert the function's actual static type).
	resolved := out.([]byte)
	return &gen.JsonLogicResult{Result: string(resolved)}, nil
}
