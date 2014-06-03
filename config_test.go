package confy

import (
	"bytes"
	"encoding/json"
	"testing"
)

var testJSON = `
{
	"Root": 5,
	"RootExtra": 10,
	"sub1": {
		"Root": 20,
		"RootExtra": 25
	},
	"sub2": {
		"Root": 30,
		"RootExtra": 35,
		"sub2.1": {
			"RootExtra": 45
		}
	},
	"sub3": {
		"Root": 100
	},
	"unrelated": "hello world"
}
`

type testConf struct {
	Root      int
	RootExtra int
}

func (tc *testConf) Default() {
	tc.Root = 9999
	tc.RootExtra = 1024
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
	if t1.Root != 20 || t1.RootExtra != 25 {
		t.Fatal("sub1 did not load correctly")
	}

	if t2.Root != 30 || t2.RootExtra != 35 {
		t.Fatal("sub2 did not load correctly")
	}

	if t21.Root != 9999 || t21.RootExtra != 45 {
		t.Fatal("sub2.1 did not load correctly")
	}

	if t3.Root != 100 || t3.RootExtra != 1024 {
		t.Fatal("sub3 did not load correctly")
	}
}

func TestCycle(t *testing.T) {
	tc := testConf{
		Root:      6000,
		RootExtra: 10000,
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
