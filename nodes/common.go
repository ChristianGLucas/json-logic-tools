package nodes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"

	jsonlogic "github.com/diegoholiveira/jsonlogic/v3"
)

// Safety bounds applied to every node, BEFORE any recursive parsing or
// evaluation is attempted. These exist to stop a native, unrecoverable Go
// stack overflow (not a panic — recover() cannot catch it) from a small but
// pathologically deep JSON document, e.g. thousands of nested single-element
// arrays easily fit inside a few KB. See checkJSONBounds.
const (
	maxLogicBytes = 256 * 1024 // 256 KiB
	maxDataBytes  = 256 * 1024 // 256 KiB
	maxJSONDepth  = 64
	evalTimeout   = 5 * time.Second
)

// checkJSONBounds performs a non-recursive, streaming scan of raw JSON bytes,
// bounding size and nesting depth before any recursive parse. It uses
// json.Decoder.Token(), which tokenizes iteratively with a simple counter
// rather than recursing — so this check itself cannot stack-overflow no
// matter how deep the (rejected) input claims to be. Must run before
// json.Unmarshal or any tree walk over the decoded value.
func checkJSONBounds(field string, raw []byte, maxBytes int) error {
	if len(raw) > maxBytes {
		return fmt.Errorf("%s exceeds max size of %d bytes", field, maxBytes)
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	depth := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("%s is not valid JSON: %w", field, err)
		}
		if d, ok := tok.(json.Delim); ok {
			switch d {
			case '{', '[':
				depth++
				if depth > maxJSONDepth {
					return fmt.Errorf("%s exceeds max nesting depth of %d", field, maxJSONDepth)
				}
			case '}', ']':
				depth--
			}
		}
	}
	return nil
}

// requireJSONField bounds-checks and decodes a required JSON field (the rule
// itself) into a generic any — the shape jsonlogic.ApplyInterface expects
// (map[string]any / []any / float64 / string / bool / nil). An empty field
// is rejected rather than defaulted, since a rule is never optional.
func requireJSONField(field, raw string, maxBytes int) (any, error) {
	if raw == "" {
		return nil, fmt.Errorf("%s is required", field)
	}
	return decodeJSONField(field, raw, maxBytes)
}

// optionalJSONField is like requireJSONField but treats an empty value as
// {} — matching the underlying library's own default for absent data.
func optionalJSONField(field, raw string, maxBytes int) (any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	return decodeJSONField(field, raw, maxBytes)
}

func decodeJSONField(field, raw string, maxBytes int) (any, error) {
	b := []byte(raw)
	if err := checkJSONBounds(field, b, maxBytes); err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON: %w", field, err)
	}
	return v, nil
}

// runWithTimeoutAndRecover runs fn on its own goroutine, bounding its
// wall-clock time AND converting any panic into a structured error. This is
// the general safety wrapper for every call into the vendored jsonlogic
// library: ApplyInterface already recovers its own panics internally, but
// at least one other entry point (GetJsonLogicWithSolvedVars, used by
// ResolveVariables) does not — a bare-literal root rule (e.g. "true") hits
// an unrecovered `rule.(map[string]any)` type assertion deep in the
// library and panics, which — unrecovered — would crash the whole process,
// not just the request. Route every library call through this wrapper
// rather than relying on the library's own (inconsistent) panic handling.
// Inputs are already bounded in size and nesting depth by checkJSONBounds
// by the time this runs, so a real timeout would indicate a library-level
// performance bug rather than a crafted input; the timeout is a defensive
// bound, not the primary safety mechanism.
func runWithTimeoutAndRecover(timeout time.Duration, fn func() (any, error)) (any, error) {
	type result struct {
		out any
		err error
	}
	ch := make(chan result, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- result{nil, fmt.Errorf("internal evaluation error: %v", r)}
			}
		}()
		out, err := fn()
		ch <- result{out, err}
	}()
	select {
	case r := <-ch:
		return r.out, r.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("evaluation exceeded the %s time limit", timeout)
	}
}

// applyWithTimeout runs jsonlogic.ApplyInterface with the standard
// timeout+panic-recovery wrapper.
func applyWithTimeout(rule, data any, timeout time.Duration) (any, error) {
	return runWithTimeoutAndRecover(timeout, func() (any, error) {
		return jsonlogic.ApplyInterface(rule, data)
	})
}

// hasNonFiniteFloat reports whether v — a value shaped like
// jsonlogic.ApplyInterface's output (float64/string/bool/nil/
// map[string]any/[]any) — contains a NaN or +/-Inf float64 anywhere.
// json.Marshal cannot encode those, and a handful of arithmetic operators
// (division/modulo by zero, among others) can produce one from otherwise
// ordinary input, so callers should check this and return a clear domain
// error instead of surfacing json.Marshal's generic encoding failure.
func hasNonFiniteFloat(v any) bool {
	switch t := v.(type) {
	case float64:
		return math.IsNaN(t) || math.IsInf(t, 0)
	case map[string]any:
		for _, vv := range t {
			if hasNonFiniteFloat(vv) {
				return true
			}
		}
	case []any:
		for _, vv := range t {
			if hasNonFiniteFloat(vv) {
				return true
			}
		}
	}
	return false
}
