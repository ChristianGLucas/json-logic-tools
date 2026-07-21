package nodes_test

import (
	"context"
	"strings"
	"testing"

	gen "christiangeorgelucas/json-logic-tools/gen"
	jsonlogic "github.com/diegoholiveira/jsonlogic/v3"

	"christiangeorgelucas/json-logic-tools/nodes"
)

func TestListOperators(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ListOperators(ctx, ax, &gen.ListOperatorsRequest{})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}

	if len(got.Operators) != 39 {
		t.Errorf("got %d operators, want 39 (see nodes/operators.go for the pinned-version cross-check)", len(got.Operators))
	}

	byName := map[string]*gen.JsonLogicOperatorInfo{}
	for _, op := range got.Operators {
		if op.Name == "" {
			t.Error("operator with empty name")
		}
		if op.Description == "" {
			t.Errorf("operator %q has no description", op.Name)
		}
		if _, dup := byName[op.Name]; dup {
			t.Errorf("duplicate operator name %q", op.Name)
		}
		byName[op.Name] = op
	}

	// Spot-check a few well-known standard operators and the non-standard
	// extensions are classified correctly.
	for _, name := range []string{"var", "if", "map", "reduce", "=="} {
		op, ok := byName[name]
		if !ok {
			t.Errorf("expected standard operator %q to be present", name)
			continue
		}
		if !op.Standard {
			t.Errorf("operator %q should be marked standard", name)
		}
	}
	for _, name := range []string{"contains_all", "contains_any", "contains_none"} {
		op, ok := byName[name]
		if !ok {
			t.Errorf("expected non-standard operator %q to be present", name)
			continue
		}
		if op.Standard {
			t.Errorf("operator %q should NOT be marked standard", name)
		}
	}
}

// TestListOperatorsMatchesLibrary is the independent-oracle test: for every
// operator this package claims to support, ask the pinned library's own
// validator (not this package's code) whether it recognizes that operator
// name. This catches a typo'd or stale entry in operators.go — it is
// intentionally independent of nodes/operators.go's own logic.
func TestListOperatorsMatchesLibrary(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ListOperators(ctx, ax, &gen.ListOperatorsRequest{})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}

	for _, op := range got.Operators {
		rule := `{` + jsonQuote(op.Name) + `:[]}`
		if !jsonlogic.IsValid(strings.NewReader(rule)) {
			t.Errorf("operator %q is listed by this package but the pinned jsonlogic library does not recognize it", op.Name)
		}
	}
}

func jsonQuote(s string) string {
	// Minimal JSON string quoting sufficient for operator names (no special
	// characters in any of them), avoiding an extra import for this test.
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
