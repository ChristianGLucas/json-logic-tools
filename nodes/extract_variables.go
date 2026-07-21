package nodes

import (
	"context"
	"sort"
	"strconv"

	"christiangeorgelucas/json-logic-tools/axiom"
	gen "christiangeorgelucas/json-logic-tools/gen"
)

// ExtractVariables statically walks a JSON Logic rule and returns the sorted,
// de-duplicated list of data variable paths it references via {"var": ...}
// — useful for figuring out what data a rule needs before running it. The
// walk mirrors the underlying library's own evaluation-time traversal
// exactly (an object with more than one key is treated as an opaque literal
// and not descended into, matching how Evaluate would treat it), so the
// paths reported are the ones Evaluate will actually look up. A variable
// name that is itself computed (e.g. {"var":{"cat":["a","b"]}}) can't be
// resolved statically and is omitted, though variables referenced inside
// that computation are still walked and reported. `data` is ignored; only
// `logic` is used. Same size/depth bounds as Evaluate.
func ExtractVariables(ctx context.Context, ax axiom.Context, input *gen.JsonLogicRule) (*gen.JsonLogicVariables, error) {
	rule, err := requireJSONField("logic", input.Logic, maxLogicBytes)
	if err != nil {
		return &gen.JsonLogicVariables{Error: err.Error()}, nil
	}

	found := map[string]struct{}{}
	walkForVars(rule, found, 0)

	vars := make([]string, 0, len(found))
	for v := range found {
		vars = append(vars, v)
	}
	sort.Strings(vars)

	return &gen.JsonLogicVariables{Variables: vars}, nil
}

// walkForVars mirrors the library's own apply()/parseValues() traversal:
// a map with exactly one key is an operator (or "var") applied to its
// value; a map with more than one key is an opaque literal and is not
// descended into; an array is walked element-by-element. depth is bounded
// by the caller's prior checkJSONBounds pass (<= maxJSONDepth), so this
// recursion is itself bounded and cannot stack-overflow.
func walkForVars(node any, found map[string]struct{}, depth int) {
	if depth > maxJSONDepth {
		return
	}
	switch v := node.(type) {
	case map[string]any:
		if len(v) != 1 {
			return
		}
		for op, val := range v {
			if op == "var" {
				recordVarName(val, found)
			}
			walkForVars(val, found, depth+1)
		}
	case []any:
		for _, el := range v {
			walkForVars(el, found, depth+1)
		}
	}
}

// recordVarName extracts a statically-known path from a {"var": ...} value,
// per the shapes the library accepts: a string path, a numeric array index,
// or a [path, default] pair (default is ignored here, but still walked by
// the caller for any vars it references). A "" path, an empty array, or a
// dynamic (map/array-computed) name means "whole data" / "unresolvable" and
// records nothing.
func recordVarName(val any, found map[string]struct{}) {
	switch t := val.(type) {
	case string:
		if t != "" {
			found[t] = struct{}{}
		}
	case float64:
		found[formatNumericIndex(t)] = struct{}{}
	case []any:
		if len(t) == 0 {
			return
		}
		switch first := t[0].(type) {
		case string:
			if first != "" {
				found[first] = struct{}{}
			}
		case float64:
			found[formatNumericIndex(first)] = struct{}{}
		}
	}
}

func formatNumericIndex(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
