package nodes

import (
	"context"

	"christiangeorgelucas/json-logic-tools/axiom"
	gen "christiangeorgelucas/json-logic-tools/gen"
)

// ListOperators returns the full catalog of JSON Logic operators this
// package's Evaluate/ResolveVariables nodes support: 36 from the official
// jsonlogic.com spec plus 3 documented non-standard extensions
// (contains_all/contains_any/contains_none), each with a short description
// and whether it's a spec-standard operator. Every operator is a pure,
// deterministic function of its arguments and the data object — no random,
// wall-clock, or external-lookup operator is registered by this package.
// Static (takes no input); TestListOperatorsMatchesLibrary independently
// verifies every name here is actually registered by the pinned library.
func ListOperators(ctx context.Context, ax axiom.Context, input *gen.ListOperatorsRequest) (*gen.JsonLogicOperators, error) {
	ops := make([]*gen.JsonLogicOperatorInfo, 0, len(operatorCatalog))
	for _, op := range operatorCatalog {
		ops = append(ops, &gen.JsonLogicOperatorInfo{
			Name:        op.Name,
			Description: op.Description,
			Standard:    op.Standard,
		})
	}
	return &gen.JsonLogicOperators{Operators: ops}, nil
}
