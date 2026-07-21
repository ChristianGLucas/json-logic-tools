package nodes

import (
	"context"
	"encoding/json"

	"christiangeorgelucas/json-logic-tools/axiom"
	gen "christiangeorgelucas/json-logic-tools/gen"
)

// Evaluate applies a JSON Logic rule (jsonlogic.com) to a data object and
// returns the result, wrapping github.com/diegoholiveira/jsonlogic/v3.
// `data` defaults to {} when omitted. `logic` and `data` are each bounded at
// 256 KiB and 64 levels of JSON nesting; a bound violation, malformed JSON,
// an unsupported operator, or a type error surfaced by an operator all come
// back as a structured error rather than a crash. Deterministic: every
// operator this package exposes is a pure function of the rule and data
// (see ListOperators) — no randomness, wall-clock, or external lookups.
func Evaluate(ctx context.Context, ax axiom.Context, input *gen.JsonLogicRule) (*gen.JsonLogicResult, error) {
	rule, err := requireJSONField("logic", input.Logic, maxLogicBytes)
	if err != nil {
		return &gen.JsonLogicResult{Error: err.Error()}, nil
	}
	data, err := optionalJSONField("data", input.Data, maxDataBytes)
	if err != nil {
		return &gen.JsonLogicResult{Error: err.Error()}, nil
	}

	out, err := applyWithTimeout(rule, data, evalTimeout)
	if err != nil {
		return &gen.JsonLogicResult{Error: err.Error()}, nil
	}

	resultJSON, err := json.Marshal(out)
	if err != nil {
		return &gen.JsonLogicResult{Error: "failed to encode result: " + err.Error()}, nil
	}

	return &gen.JsonLogicResult{Result: string(resultJSON)}, nil
}
