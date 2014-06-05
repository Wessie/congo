package congo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

var testJSON = `
{
	"Root": 5,
	"Extra": 10,
	"sub1": {
		"Root": 15,
		"Extra": 20
	},
	"sub2": {
		"Root": 25,
		"Extra": 30,
		"sub2.1": {
			"Extra": 35
		}
	},
	"unrelated": "hello world"
}
`

const (
	tcRootDefault  = 9999
	tcExtraDefault = 1024
)

type testConf struct {
	Root  int
	Extra int
}

func (tc *testConf) Default() {
	tc.Root = tcRootDefault
	tc.Extra = tcExtraDefault
}

func TestNested(t *testing.T) {
	c := NewConfig(&testConf{})

	t1, t2, t21, t3 := &testConf{}, &testConf{}, &testConf{}, &testConf{}

	c.AddSub("sub1", t1)
	c.AddSub("sub2", t2)
	c.Subs["sub2"].AddSub("sub2.1", t21)
	c.AddSub("sub3", t3)

	if err := json.Unmarshal([]byte(testJSON), c); err != nil {
		t.Fatal(err)
	}

	// TODO: Improve error messages
	if t1.Root != 15 || t1.Extra != 20 {
		t.Errorf("sub1 did not load correctly: %v != 15 or/and %v != 20", t1.Root, t1.Extra)
	}

	if t2.Root != 25 || t2.Extra != 30 {
		t.Errorf("sub2 did not load correctly: %v != 25 or/and %v != 30", t2.Root, t2.Extra)
	}

	if t21.Root != tcRootDefault || t21.Extra != 35 {
		t.Errorf("sub2.1 did not load correctly: %v != %v or/and %v != 35",
			t21.Root, tcRootDefault, t21.Extra)
	}

	if t3.Root != tcRootDefault || t3.Extra != tcExtraDefault {
		t.Errorf("sub3 did not load correctly: %v != %v or/and %v != %v",
			t3.Root, tcRootDefault, t3.Extra, tcExtraDefault)
	}
}

func TestCycle(t *testing.T) {
	tc := testConf{
		Root:  6000,
		Extra: 10000,
	}
	c := NewConfig(&tc)

	b := new(bytes.Buffer)

	if err := c.SaveWriter(b); err != nil {
		t.Fatal("failed saving to writer:", err)
	}

	tk := testConf{}
	k := NewConfig(&tk)

	if err := k.LoadReader(b); err != nil {
		t.Fatal("failed loading from reader:", err)
	}

	if tk != tc {
		t.Fatalf("cycling configuration did not return original: %+v != %+v", tk, tc)
	}
}

func TestDefault(t *testing.T) {
	tc := &testConf{}
	c := NewConfig(tc)

	if err := c.LoadReader(strings.NewReader("")); err != nil {
		t.Fatal("failed loading from empty reader:", err)
	}

	if tc.Root != tcRootDefault {
		t.Errorf("default value was not set correctly: %v != %v", tc.Root, tcRootDefault)
	}
	if tc.Extra != tcExtraDefault {
		t.Errorf("default value was not set correctly: %v != %v", tc.Extra, tcExtraDefault)
	}
}

func TestNilRoot(t *testing.T) {
	tc, c := new(testConf), NewConfig(nil)
	c.AddSub("sub", tc)

	conf := `{"sub":{"Root":50}}`
	if err := c.UnmarshalJSON([]byte(conf)); err != nil {
		t.Fatal("failed unmarshalling nil root:", err)
	}

	if b, err := c.MarshalJSON(); err != nil {
		t.Fatal("failed marshalling nil root:", err)
	} else if string(b) != conf {
		t.Fatal("received different result: %v != %v", string(b), conf)
	}
}
