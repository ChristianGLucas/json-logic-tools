package nodes

import (
	"context"
	"fmt"

	"christiangeorgelucas/json-logic-tools/axiom"
	gen "christiangeorgelucas/json-logic-tools/gen"
	jsonlogic "github.com/diegoholiveira/jsonlogic/v3"
)

// ValidateRule checks whether a rule is syntactically valid JSON Logic —
// valid JSON, and every operator used (recursively) is one this package
// supports (see ListOperators) — wrapping the library's own
// ValidateJsonLogic. This does not evaluate the rule or check it against
// data, so a structurally valid rule can still error at Evaluate time (e.g.
// a type mismatch an operator surfaces at runtime). `data` is ignored.
// Same size/depth bounds as Evaluate; a bound violation is reported as
// valid=false with a structured reason, never a crash.
func ValidateRule(ctx context.Context, ax axiom.Context, input *gen.JsonLogicRule) (*gen.ValidateRuleResult, error) {
	rule, err := requireJSONField("logic", input.Logic)
	if err != nil {
		return &gen.ValidateRuleResult{Valid: false, Error: err.Error()}, nil
	}

	if jsonlogic.ValidateJsonLogic(rule) {
		return &gen.ValidateRuleResult{Valid: true}, nil
	}

	if bad := firstUnsupportedOperator(rule, 0); bad != "" {
		return &gen.ValidateRuleResult{
			Valid: false,
			Error: fmt.Sprintf("unsupported operator: %q", bad),
		}, nil
	}

	return &gen.ValidateRuleResult{
		Valid: false,
		Error: "rule is not a well-formed JSON Logic structure",
	}, nil
}

// firstUnsupportedOperator walks the same shape as walkForVars, looking for
// the first operator key not in knownOperators, to give ValidateRule a more
// actionable reason than the library's plain bool. Returns "" if none is
// found structurally (e.g. the failure is some other malformed shape).
func firstUnsupportedOperator(node any, depth int) string {
	if depth > maxJSONDepth {
		return ""
	}
	switch v := node.(type) {
	case map[string]any:
		if len(v) != 1 {
			return ""
		}
		for op, val := range v {
			if op != "var" {
				if _, ok := knownOperators[op]; !ok {
					return op
				}
			}
			if bad := firstUnsupportedOperator(val, depth+1); bad != "" {
				return bad
			}
		}
	case []any:
		for _, el := range v {
			if bad := firstUnsupportedOperator(el, depth+1); bad != "" {
				return bad
			}
		}
	}
	return ""
}
