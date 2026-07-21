package nodes_test

import (
	"context"
	"strings"
	"testing"

	gen "christiangeorgelucas/json-logic-tools/gen"
	"christiangeorgelucas/json-logic-tools/nodes"
)

func TestValidateRule(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	cases := []struct {
		name       string
		logic      string
		wantValid  bool
		errorMatch string // substring the error must contain when wantValid is false; "" = don't check
	}{
		{"valid simple equality", `{"==":[{"var":"a"},1]}`, true, ""},
		{"valid nested rule", `{"and":[{"<":[{"var":"temp"},110]},{"!=":[{"var":"pie.filling"},"apple"]}]}`, true, ""},
		{"valid literal", `true`, true, ""},
		{"malformed JSON", `{"==": [1, `, false, "not valid JSON"},
		{"unsupported operator", `{"definitely_not_an_operator":[1,2]}`, false, "definitely_not_an_operator"},
		{"unsupported operator nested", `{"and":[true,{"nope":[1]}]}`, false, "nope"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := nodes.ValidateRule(ctx, ax, &gen.JsonLogicRule{Logic: tc.logic})
			if err != nil {
				t.Fatalf("unexpected transport error: %v", err)
			}
			if got.Valid != tc.wantValid {
				t.Errorf("ValidateRule(%s).Valid = %v, want %v (error=%q)", tc.logic, got.Valid, tc.wantValid, got.Error)
			}
			if !tc.wantValid {
				if got.Error == "" {
					t.Error("expected a non-empty error reason for an invalid rule")
				}
				if tc.errorMatch != "" && !strings.Contains(got.Error, tc.errorMatch) {
					t.Errorf("error = %q, want it to contain %q", got.Error, tc.errorMatch)
				}
			} else if got.Error != "" {
				t.Errorf("expected empty error for a valid rule, got %q", got.Error)
			}
		})
	}
}

func TestValidateRuleErrorPaths(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	t.Run("empty logic", func(t *testing.T) {
		got, err := nodes.ValidateRule(ctx, ax, &gen.JsonLogicRule{Logic: ""})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Valid {
			t.Fatal("empty logic should not be valid")
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for empty logic")
		}
	})

	t.Run("nested beyond max depth", func(t *testing.T) {
		deep := strings.Repeat(`[`, 100) + `1` + strings.Repeat(`]`, 100)
		got, err := nodes.ValidateRule(ctx, ax, &gen.JsonLogicRule{Logic: deep})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Valid {
			t.Fatal("excessively deep logic should not be valid")
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for excessive nesting depth")
		}
	})
}
