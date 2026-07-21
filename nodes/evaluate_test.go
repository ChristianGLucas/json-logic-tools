package nodes_test

import (
	"context"
	"strings"
	"testing"

	gen "christiangeorgelucas/json-logic-tools/gen"
	"christiangeorgelucas/json-logic-tools/nodes"
)

// Every case here is an independent-oracle test: the (rule, data) -> result
// pair is taken directly from the official jsonlogic.com documentation or
// from the pinned diegoholiveira/jsonlogic v3.10.1 test suite (lists_test.go)
// — not derived from this package's own implementation.
func TestEvaluateJsonLogicComExamples(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	cases := []struct {
		name  string
		logic string
		data  string
		want  string // JSON-encoded expected result
	}{
		{"simple equality", `{"==":[{"var":"a"},1]}`, `{"a":1}`, `true`},
		{"var lookup", `{"var":"a"}`, `{"a":1}`, `1`},
		{"var with default sugar", `{"var":["a",1]}`, `{}`, `1`},
		{"var dot path", `{"var":"champ.name"}`, `{"champ":{"name":"Sadio Mane"}}`, `"Sadio Mane"`},
		{"if/then/else", `{"if":[true,"yes","no"]}`, `{}`, `"yes"`},
		{"addition", `{"+":[1,2]}`, `{}`, `3`},
		{"between form", `{"<":[1,2,3]}`, `{}`, `true`},
		{"filter", `{"filter":[{"var":"integers"},{">":[{"var":""},2]}]}`, `{"integers":[1,2,3,4,5]}`, `[3,4,5]`},
		{"map", `{"map":[{"var":"integers"},{"*":[{"var":""},2]}]}`, `{"integers":[1,2,3]}`, `[2,4,6]`},
		{"reduce sum", `{"reduce":[{"var":"integers"},{"+":[{"var":"current"},{"var":"accumulator"}]},0]}`, `{"integers":[1,2,3,4]}`, `10`},
		{"reduce skips null", `{"reduce":[[1,2,null,4,5],{"+":[{"var":"current"},{"var":"accumulator"}]},0]}`, `{}`, `12`},
		{"in array", `{"in":["Ringo",["John","Paul","George","Ringo"]]}`, `{}`, `true`},
		{"cat", `{"cat":["I love ","pie"]}`, `{}`, `"I love pie"`},
		{"missing", `{"missing":["a","b"]}`, `{"a":"apple"}`, `["b"]`},
		{"non-standard contains_all", `{"contains_all":[["a","b","c"],["a","b"]]}`, `{}`, `true`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := nodes.Evaluate(ctx, ax, &gen.JsonLogicRule{Logic: tc.logic, Data: tc.data})
			if err != nil {
				t.Fatalf("unexpected transport error: %v", err)
			}
			if got.Error != "" {
				t.Fatalf("unexpected node error: %s", got.Error)
			}
			if strings.TrimSpace(got.Result) != tc.want {
				t.Errorf("Evaluate(%s, %s) = %s, want %s", tc.logic, tc.data, got.Result, tc.want)
			}
		})
	}
}

func TestEvaluateIsDeterministic(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	input := &gen.JsonLogicRule{
		Logic: `{"reduce":[{"var":"integers"},{"+":[{"var":"current"},{"var":"accumulator"}]},0]}`,
		Data:  `{"integers":[1,2,3,4,5,6,7,8,9,10]}`,
	}

	first, err := nodes.Evaluate(ctx, ax, input)
	if err != nil || first.Error != "" {
		t.Fatalf("first invocation failed: err=%v nodeErr=%s", err, first.Error)
	}
	second, err := nodes.Evaluate(ctx, ax, input)
	if err != nil || second.Error != "" {
		t.Fatalf("second invocation failed: err=%v nodeErr=%s", err, second.Error)
	}
	if first.Result != second.Result {
		t.Errorf("non-deterministic: first=%s second=%s", first.Result, second.Result)
	}
	if first.Result != "55" {
		t.Errorf("got %s, want 55", first.Result)
	}
}

func TestEvaluateErrorPaths(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	t.Run("empty logic", func(t *testing.T) {
		got, err := nodes.Evaluate(ctx, ax, &gen.JsonLogicRule{Logic: ""})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for empty logic")
		}
		if got.Result != "" {
			t.Errorf("result should be empty on error, got %q", got.Result)
		}
	})

	t.Run("malformed logic JSON", func(t *testing.T) {
		got, err := nodes.Evaluate(ctx, ax, &gen.JsonLogicRule{Logic: `{"==": [1, `})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for malformed JSON")
		}
	})

	t.Run("malformed data JSON", func(t *testing.T) {
		got, err := nodes.Evaluate(ctx, ax, &gen.JsonLogicRule{Logic: `{"var":"a"}`, Data: `not json`})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for malformed data JSON")
		}
	})

	t.Run("unsupported operator", func(t *testing.T) {
		got, err := nodes.Evaluate(ctx, ax, &gen.JsonLogicRule{Logic: `{"definitely_not_an_operator":[1,2]}`})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for an unsupported operator")
		}
	})

	t.Run("oversized logic", func(t *testing.T) {
		huge := `{"var":"` + strings.Repeat("a", 300*1024) + `"}`
		got, err := nodes.Evaluate(ctx, ax, &gen.JsonLogicRule{Logic: huge})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for oversized logic")
		}
	})

	t.Run("logic nested beyond max depth", func(t *testing.T) {
		// 100 nested single-element arrays — small in bytes, deep in structure.
		deep := strings.Repeat(`[`, 100) + `1` + strings.Repeat(`]`, 100)
		got, err := nodes.Evaluate(ctx, ax, &gen.JsonLogicRule{Logic: deep})
		if err != nil {
			t.Fatalf("unexpected transport error: %v", err)
		}
		if got.Error == "" {
			t.Fatal("expected a structured error for excessive nesting depth")
		}
	})
}
