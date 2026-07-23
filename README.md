# json-logic-tools

Composable [Axiom](https://axiomide.com) nodes for evaluating [JSON Logic](https://jsonlogic.com)
rules — portable, JSON-encoded business/decision rules — wrapping the
MIT-licensed [`github.com/diegoholiveira/jsonlogic/v3`](https://github.com/diegoholiveira/jsonlogic)
(pinned at v3.10.1).

Built for the Axiom marketplace, handle `christiangeorgelucas`.

## Use it from your agent or app

Every node in this package is a **live, auto-scaling API endpoint** on the
[Axiom](https://axiomide.com) marketplace — call it from an AI agent or your own
code, with nothing to self-host.

**📦 See it on the marketplace:**
https://dev.axiomide.com/marketplace/christiangeorgelucas/json-logic-tools@0.1.0

**Hook it up to an AI agent (MCP).** Add Axiom's hosted MCP server to any MCP
client and every node becomes a typed tool your agent can call — search the
catalog, inspect a schema, and invoke it directly.

```bash
# Claude Code
claude mcp add --transport http axiom https://api.axiomide.com/mcp \
  --header "Authorization: Bearer $AXIOM_API_KEY"
```

Claude Desktop, Cursor, or any config-based client:

```json
{
  "mcpServers": {
    "axiom": {
      "type": "http",
      "url": "https://api.axiomide.com/mcp",
      "headers": { "Authorization": "Bearer YOUR_AXIOM_API_KEY" }
    }
  }
}
```

**Call it from the CLI.**

```bash
axiom invoke christiangeorgelucas/json-logic-tools/Evaluate --input '{ ... }'
```

**Call it over HTTP.**

```bash
curl -X POST https://api.axiomide.com/invocations/v1/nodes/christiangeorgelucas/json-logic-tools/0.1.0/Evaluate \
  -H "Authorization: Bearer $AXIOM_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ ... }'
```

> Input/output schema for each node is on the marketplace page above, or via
> `axiom inspect node christiangeorgelucas/json-logic-tools/Evaluate`.

### Get started free

Install the CLI:

```bash
# macOS / Linux — Homebrew
brew install axiomide/tap/axiom

# macOS / Linux — install script
curl -fsSL https://raw.githubusercontent.com/AxiomIDE/axiom-releases/main/install.sh | sh
```

**Windows:** download the `windows/amd64` `.zip` from the
[releases page](https://github.com/AxiomIDE/axiom-releases/releases), unzip it,
and put `axiom.exe` on your `PATH`.

Then `axiom version` to verify, `axiom login` (GitHub or Google) to authenticate,
and create an API key under **Console → API Keys**. Docs and sign-up at
**[axiomide.com](https://axiomide.com)**.

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
