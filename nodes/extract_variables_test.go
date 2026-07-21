package nodes_test

import (
	"context"
	"reflect"
	"testing"

	gen "christiangeorgelucas/json-logic-tools/gen"
	"christiangeorgelucas/json-logic-tools/nodes"
)

func TestExtractVariables(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	cases := []struct {
		name  string
		logic string
		want  []string
	}{
		{"single var", `{"==":[{"var":"a"},1]}`, []string{"a"}},
		{"two vars across and/comparisons", `{"and":[{"<":[{"var":"temp"},110]},{"==":[{"var":"pie.filling"},"apple"]}]}`, []string{"pie.filling", "temp"}},
		{"root reference is not a named variable", `{"var":""}`, []string{}},
		{"empty array root reference", `{"var":[]}`, []string{}},
		{"numeric index var", `{"var":1}`, []string{"1"}},
		{"var with default sugar", `{"var":["a",1]}`, []string{"a"}},
		{"relative var inside map", `{"map":[{"var":"items"},{"var":".price"}]}`, []string{".price", "items"}},
		{"no vars at all", `{"+":[1,2]}`, []string{}},
		{"multi-key object is opaque literal, not walked", `{"==":[{"var":"a"},{"x":{"var":"should_not_appear"},"y":2}]}`, []string{"a"}},
		{"dynamic var name is unresolvable but nested vars still found", `{"var":{"cat":[{"var":"prefix"},"_id"]}}`, []string{"prefix"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := nodes.ExtractVariables(ctx, ax, &gen.JsonLogicRule{Logic: tc.logic})
			if err != nil {
				t.Fatalf("unexpected transport error: %v", err)
			}
			if got.Error != "" {
				t.Fatalf("unexpected node error: %s", got.Error)
			}
			gotVars := got.Variables
			if gotVars == nil {
				gotVars = []string{}
			}
			if !reflect.DeepEqual(gotVars, tc.want) {
				t.Errorf("ExtractVariables(%s) = %v, want %v", tc.logic, gotVars, tc.want)
			}
		})
	}
}

func TestExtractVariablesErrorPaths(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	t.Run("empty logic", func(t *testing.T) {
		got, err := nodes.ExtractVariables(ctx, ax, &gen.JsonLogicRule{Logic: ""})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for empty logic")
		}
	})

	t.Run("malformed logic", func(t *testing.T) {
		got, err := nodes.ExtractVariables(ctx, ax, &gen.JsonLogicRule{Logic: `[1,2,`})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for malformed logic JSON")
		}
	})
}
