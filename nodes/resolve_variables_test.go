package nodes_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	gen "christiangeorgelucas/json-logic-tools/gen"
	"christiangeorgelucas/json-logic-tools/nodes"
)

func jsonEqual(t *testing.T, got, want string) bool {
	t.Helper()
	var g, w any
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Fatalf("got is not valid JSON: %v (%s)", err, got)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("want is not valid JSON: %v (%s)", err, want)
	}
	gb, _ := json.Marshal(g)
	wb, _ := json.Marshal(w)
	return string(gb) == string(wb)
}

// TestResolveVariablesLibraryTestSuiteExample is the independent-oracle
// test: this exact (logic, data) -> resolved triple is taken verbatim from
// the pinned diegoholiveira/jsonlogic v3.10.1 test suite
// (TestSolveVarsBackToJsonLogicWithUnicodeChars in vars_test.go), which
// exercises GetJsonLogicWithSolvedVars directly — the function this node
// wraps. (The library's own README shows a second example whose comment
// claims a boolean-coercing result inconsistent with what the pinned
// version actually returns for a string data value — the test suite, not
// that comment, is the oracle trusted here.)
func TestResolveVariablesLibraryTestSuiteExample(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ResolveVariables(ctx, ax, &gen.JsonLogicRule{
		Logic: `{">=":[{"var":"value"},10]}`,
		Data:  `{"value":20}`,
	})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	want := `{">=":[20,10]}`
	if !jsonEqual(t, got.Result, want) {
		t.Errorf("ResolveVariables = %s, want %s", got.Result, want)
	}
}

func TestResolveVariablesDefaultsDataToEmptyObject(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ResolveVariables(ctx, ax, &gen.JsonLogicRule{
		Logic: `{"var":["a","fallback"]}`,
	})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}
	if !jsonEqual(t, got.Result, `{"var":["a","fallback"]}`) {
		t.Errorf("ResolveVariables = %s, want unresolved var (default data {})", got.Result)
	}
}

func TestResolveVariablesErrorPaths(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	t.Run("empty logic", func(t *testing.T) {
		got, err := nodes.ResolveVariables(ctx, ax, &gen.JsonLogicRule{Logic: ""})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for empty logic")
		}
	})

	t.Run("malformed logic", func(t *testing.T) {
		got, err := nodes.ResolveVariables(ctx, ax, &gen.JsonLogicRule{Logic: `{not json`})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for malformed logic JSON")
		}
	})

	t.Run("large data resolves without a size cap", func(t *testing.T) {
		// No payload size cap anymore (the platform owns size); large but
		// well-formed data must resolve cleanly, not be rejected or crash.
		huge := `{"a":"` + strings.Repeat("x", 300*1024) + `"}`
		got, err := nodes.ResolveVariables(ctx, ax, &gen.JsonLogicRule{Logic: `{"var":"a"}`, Data: huge})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error != "" {
			t.Fatalf("large data should resolve, got structured error: %s", got.Error)
		}
	})

	// Regression test for a CRITICAL finding from independent review: the
	// vendored library's GetJsonLogicWithSolvedVars does an unrecovered
	// `rule.(map[string]any)` type assertion on the root value, which
	// panics — and, unrecovered, crashes the whole process, not just this
	// request — on a bare literal or array root. ValidateRule's own test
	// suite confirms these ARE valid JSON Logic rules (a literal evaluates
	// to itself), so this is a realistic input, not a contrived one: any
	// caller who validates a rule then resolves it can hit this exact
	// shape. Every case below must come back as a clean node-level error,
	// and critically must not panic/crash the test process.
	for _, tc := range []struct {
		name  string
		logic string
	}{
		{"bare boolean literal", `true`},
		{"bare number literal", `42`},
		{"bare string literal", `"hello"`},
		{"bare array root", `[1,2,3]`},
		{"bare null literal", `null`},
	} {
		t.Run("root shape rejected cleanly: "+tc.name, func(t *testing.T) {
			got, err := nodes.ResolveVariables(ctx, ax, &gen.JsonLogicRule{Logic: tc.logic})
			if err != nil {
				t.Fatalf("unexpected transport error: %v", err)
			}
			if got.Error == "" {
				t.Fatalf("expected a structured error for a %s root, got result=%q", tc.name, got.Result)
			}
		})
	}
}
