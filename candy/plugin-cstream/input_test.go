package cstream

import (
	"testing"

	"github.com/opencharly/plugin-cstream/candy/plugin-cstream/params"
)

// Each input method renders exactly the cstream-input action its fields describe,
// and a method missing its field is an authoring error, never an empty send.
func TestInputArgs(t *testing.T) {
	ok := []struct {
		in   params.CstreamInput
		want string
	}{
		{params.CstreamInput{Method: "type", Text: "touch /tmp/x"}, "type=touch /tmp/x"},
		{params.CstreamInput{Method: "key", Key: "Return"}, "key=Return"},
		{params.CstreamInput{Method: "move", X: 10.5, Y: 20}, "move=10.5,20"},
		{params.CstreamInput{Method: "click", X: 1, Y: 2}, "click=1,2,1"},
		{params.CstreamInput{Method: "click", X: 1, Y: 2, Button: 3}, "click=1,2,3"},
		{params.CstreamInput{Method: "scroll", X: 5, Y: 6, DeltaY: -120}, "scroll=5,6,0,-120"},
	}
	for _, c := range ok {
		got, err := inputArgs(&c.in)
		if err != nil || got != c.want {
			t.Errorf("%+v -> %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	for _, bad := range []params.CstreamInput{
		{Method: "type"},
		{Method: "key"},
		{Method: "scroll", X: 1, Y: 1},
		{Method: "status"},
	} {
		if got, err := inputArgs(&bad); err == nil {
			t.Errorf("%+v rendered %q; want an error", bad, got)
		}
	}
}

// The action reaches the venue as ONE shell word, whatever the typed text holds.
func TestTypedTextIsOneShellWord(t *testing.T) {
	got := shellQuote("type=echo 'hi'; rm -rf /")
	want := `'type=echo '\''hi'\''; rm -rf /'`
	if got != want {
		t.Fatalf("shellQuote = %s, want %s", got, want)
	}
}
