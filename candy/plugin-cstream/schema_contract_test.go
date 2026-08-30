package cstream

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/opencharly/plugin-cstream/candy/plugin-cstream/params"
)

// The CUE schema is the single source (SDD): it generates params/ AND is served
// over Describe so the host validates every authored step against it. Nothing
// enforces that the Go side keeps up, so these tests do.
//
// This matters here specifically: `frame` was withheld from the enum while the
// media path was broken, and re-adding it touches three places (the enum, the
// generated struct, the dispatch switch). A method present in one and missing
// from another fails at bed time as `unknown method`, or worse, validates and
// then does nothing.

var methodEnumRe = regexp.MustCompile(`(?m)^\s*method:\s*(.+)$`)

func schemaMethods(t *testing.T) []string {
	t.Helper()
	src, err := os.ReadFile("schema/cstream.cue")
	if err != nil {
		t.Fatalf("reading the CUE schema: %v", err)
	}
	m := methodEnumRe.FindSubmatch(src)
	if m == nil {
		t.Fatal("no `method:` enum found in schema/cstream.cue")
	}
	var out []string
	for _, part := range strings.Split(string(m[1]), "|") {
		out = append(out, strings.Trim(strings.TrimSpace(part), `"`))
	}
	sort.Strings(out)
	return out
}

// TestDispatchCoversEverySchemaMethod fails when the CUE enum offers a method the
// dispatch switch does not serve. Authoring it would pass host validation and then
// die at runtime on the default arm.
func TestDispatchCoversEverySchemaMethod(t *testing.T) {
	src, err := os.ReadFile("methods.go")
	if err != nil {
		t.Fatalf("reading methods.go: %v", err)
	}
	body := string(src)
	for _, m := range schemaMethods(t) {
		if !strings.Contains(body, `case "`+m+`":`) {
			t.Errorf("schema offers method %q but dispatch has no `case %q:` — "+
				"an authored step would validate and then fail as unknown method", m, m)
		}
	}
}

// TestSchemaOffersEveryDispatchedMethod is the converse: a method served but not
// declared is unreachable, because the host validates the step against the CUE
// def before the provider ever sees it.
func TestSchemaOffersEveryDispatchedMethod(t *testing.T) {
	src, err := os.ReadFile("methods.go")
	if err != nil {
		t.Fatalf("reading methods.go: %v", err)
	}
	declared := map[string]bool{}
	for _, m := range schemaMethods(t) {
		declared[m] = true
	}
	caseRe := regexp.MustCompile(`(?m)^\s*case "([a-z-]+)":`)
	for _, m := range caseRe.FindAllStringSubmatch(string(src), -1) {
		if !declared[m[1]] {
			t.Errorf("dispatch serves method %q but the CUE enum does not offer it — "+
				"the host rejects the step before the provider is reached", m[1])
		}
	}
}

// TestGeneratedParamsMatchSchemaFields guards the generated struct against the
// schema drifting out from under it. `frame` is useless without `artifact`: the
// provider gates artifact validation on in.Artifact != "", so a struct missing
// the field silently downgrades the check to "did not error".
func TestGeneratedParamsMatchSchemaFields(t *testing.T) {
	src, err := os.ReadFile("schema/cstream.cue")
	if err != nil {
		t.Fatalf("reading the CUE schema: %v", err)
	}
	fieldRe := regexp.MustCompile(`(?m)^\s*([a-z_]+)\??:\s`)
	inSchema := map[string]bool{}
	for _, m := range fieldRe.FindAllStringSubmatch(string(src), -1) {
		inSchema[m[1]] = true
	}
	rt := reflect.TypeOf(params.CstreamInput{})
	inGo := map[string]bool{}
	for i := 0; i < rt.NumField(); i++ {
		tag := rt.Field(i).Tag.Get("json")
		inGo[strings.Split(tag, ",")[0]] = true
	}
	for name := range inSchema {
		if !inGo[name] {
			t.Errorf("schema declares %q but the generated struct has no such json field — regenerate params/", name)
		}
	}
	for name := range inGo {
		if !inSchema[name] {
			t.Errorf("generated struct carries %q which the schema does not declare — params/ is stale", name)
		}
	}
}
