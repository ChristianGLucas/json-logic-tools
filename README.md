# json-logic-tools

Composable [Axiom](https://axiom.co) nodes for evaluating [JSON Logic](https://jsonlogic.com)
rules — portable, JSON-encoded business/decision rules — wrapping the
MIT-licensed [`github.com/diegoholiveira/jsonlogic/v3`](https://github.com/diegoholiveira/jsonlogic)
(pinned at v3.10.1).

Built for the Axiom marketplace, handle `christiangeorgelucas`.

## Nodes

- **Evaluate** — apply a rule to a data object and return the result. E.g.
  `{"==":[{"var":"a"},1]}` with `{"a":1}` → `true`.
- **ResolveVariables** — partially evaluate a rule: substitute every
  `{"var": ...}` reference with its value from data, without evaluating any
  operator. Useful for auditing what a rule will see.
- **ExtractVariables** — statically walk a rule and list every data variable
  path it references, without running it.
- **ValidateRule** — check that a rule is syntactically well-formed JSON
  Logic (valid JSON, every operator recognized), without evaluating it.
- **ListOperators** — return the full catalog of supported operators (36
  from the official jsonlogic.com spec, plus 3 documented non-standard
  extensions: `contains_all`/`contains_any`/`contains_none`), each with a
  description.

## Design

Every node shares a single `JsonLogicRule` input envelope (`logic` + `data`,
both JSON text) and a `JsonLogicResult`/`JsonLogicVariables`/
`ValidateRuleResult` output. All operators exposed are pure, deterministic
functions of their arguments and the data object — no randomness,
wall-clock, or external lookups. Rule and data JSON are each bounded in size
(256 KiB) and nesting depth (64 levels) *before* any parsing is attempted,
so a small-but-pathologically-deep input is rejected with a structured
error instead of crashing the process. Stateless and fully offline.

## License

MIT — see [LICENSE](./LICENSE).
