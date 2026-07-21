package nodes

// operatorCatalog is the full, hand-cross-checked list of operators
// registered by the pinned github.com/diegoholiveira/jsonlogic/v3 v3.10.1
// (see operation.go's init() in that module) — 36 operators from the
// official jsonlogic.com spec plus 3 documented non-standard extensions
// (contains_all/contains_any/contains_none). All are pure, deterministic
// functions of their own arguments and the data object: no random, no
// wall-clock/"now", no I/O, no external lookups.
//
// TestListOperatorsMatchesLibrary (list_operators_test.go) is the
// independent-oracle check that every name below is actually registered by
// the library, by asking the library's own validator (jsonlogic.IsValid)
// whether {"<name>": []} is a recognized operator form. That test guards
// against a typo'd or stale *name* in this list; it cannot by itself prove
// the list is exhaustive (the registry is unexported), so completeness rests
// on the source cross-check above and is pinned to v3.10.1 by go.mod.
var operatorCatalog = []struct {
	Name        string
	Description string
	Standard    bool
}{
	{"var", "Retrieve a value from the data object by dot-path (\"a.b\"), numeric array index, or a default when the path is missing.", true},
	{"missing", "Given one or more data keys, returns the array of keys that are missing (or null/undefined) in the data.", true},
	{"missing_some", "Takes a minimum-required count and an array of keys; returns the missing keys, unless at least that many keys are present.", true},
	{"if", "Standard if/then/elseif.../else: takes an odd number of arguments (condition, value, [condition, value,]... default).", true},
	{"?:", "Alias for \"if\" restricted to exactly three arguments (condition, then, else).", true},
	{"==", "Loose equality, with JavaScript-style type coercion.", true},
	{"===", "Strict equality: true only if both value and type match.", true},
	{"!=", "Loose inequality (negation of ==).", true},
	{"!==", "Strict inequality (negation of ===).", true},
	{"!", "Logical NOT, using JavaScript truthiness coercion.", true},
	{"!!", "Cast to boolean using JavaScript truthiness coercion (double negation).", true},
	{"or", "Returns the first truthy argument, or the last argument if none are truthy; short-circuits.", true},
	{"and", "Returns the first falsy argument, or the last argument if all are truthy; short-circuits.", true},
	{"<", "Less than; with three arguments, tests the between form a < b < c.", true},
	{"<=", "Less than or equal; with three arguments, tests the between form a <= b <= c.", true},
	{">", "Greater than.", true},
	{">=", "Greater than or equal.", true},
	{"max", "Maximum of the given numeric arguments.", true},
	{"min", "Minimum of the given numeric arguments.", true},
	{"+", "Sum of the given arguments (numeric cast of a single argument).", true},
	{"-", "Difference of the given arguments (unary negation of a single argument).", true},
	{"*", "Product of the given arguments.", true},
	{"/", "Quotient of the given arguments.", true},
	{"%", "Remainder (modulo) of the given arguments.", true},
	{"abs", "Absolute value of a numeric argument.", true},
	{"map", "Apply a rule to every element of an array, collecting the results into a new array.", true},
	{"reduce", "Fold an array to a single value: a rule combines a running accumulator with each element in turn.", true},
	{"filter", "Keep only the elements of an array for which a rule is truthy.", true},
	{"all", "True iff every element of an array satisfies a rule (false for an empty array).", true},
	{"none", "True iff no element of an array satisfies a rule (true for an empty array).", true},
	{"some", "True iff at least one element of an array satisfies a rule.", true},
	{"merge", "Flatten one or more arrays (and scalars) into a single array.", true},
	{"in", "True if a value is found in an array, or a substring is found in a string.", true},
	{"cat", "Concatenate all arguments into a single string.", true},
	{"substr", "Extract a substring by start index and optional length; negative indices count from the end.", true},
	{"set", "Return a copy of an object/array with one property set to a value, addressed by dot-path.", true},
	{"contains_all", "Non-standard extension: true iff every element of the second array is present in the first.", false},
	{"contains_any", "Non-standard extension: true iff at least one element of the second array is present in the first.", false},
	{"contains_none", "Non-standard extension: true iff no element of the second array is present in the first.", false},
}

// knownOperators is operatorCatalog's name set, for fast membership checks
// (used by the rule-structure walk in extract_variables.go / validate_rule.go).
var knownOperators = func() map[string]struct{} {
	m := make(map[string]struct{}, len(operatorCatalog))
	for _, op := range operatorCatalog {
		m[op.Name] = struct{}{}
	}
	return m
}()
