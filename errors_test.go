package congo

import (
	"encoding/json"
	"strings"
	"testing"
)

var (
	syntaxError = `{"welcome": 5"}`  // invalid position of quote
	eofError    = `{"welcome": "5}`  // unclosed string
	typeError   = `{"welcome": "5"}` // string instead of int
)

func TestSyntaxError(t *testing.T) {
	c := NewConfig(nil)

	err := c.LoadReader(strings.NewReader(syntaxError))
	if _, ok := err.(*json.SyntaxError); err == nil {
		t.Error("expected non-nil error:", syntaxError)
	} else if !ok {
		t.Errorf("expected syntax error, got: (%T) %s", err, err)
	}

	t.Log("normal error:", err)
	t.Log("pretty error:", PrettifyError(err))
}

func TestUnexpectedEOFError(t *testing.T) {
	c := NewConfig(nil)

	err := c.LoadReader(strings.NewReader(eofError))
	if err == nil {
		t.Error("expected non-nil error:", eofError)
	}

	t.Log("normal error:", err)
	t.Log("pretty error:", PrettifyError(err))
}

func TestTypeError(t *testing.T) {
	type tc struct {
		Welcome int
	}

	c := NewConfig(&tc{})

	err := c.LoadReader(strings.NewReader(typeError))
	if err == nil {
		t.Error("expected non-nil error:", typeError)
	} else if _, ok := err.(*json.UnmarshalTypeError); !ok {
		t.Errorf("expected UnmarshalTypeError, got: (%T) %s", err, err)
	}

	t.Log("normal error:", err)
	t.Log("pretty error:", PrettifyError(err))
}
