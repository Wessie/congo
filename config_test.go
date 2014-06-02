package confy

import (
	"encoding/json"
	"testing"
)

var testJSON = `
{
	"root": 5,
	"root_extra": 10,
	"sub1": {
		"root": 20,
		"root_extra": 25
	},
	"sub2": {
		"root": 30,
		"root_extra": 35,
		"sub2.1": {
			"root_extra": 45
		}
	},
	"sub3": {
		"root": 100
	},
	"unrelated": "hello world"
}
`

type testConf struct {
	Root      int `json:"root"`
	RootExtra int `json:"root_extra"`
}

func (tc *testConf) Default() {
	tc.Root = 9999
	tc.RootExtra = 1024
}

func TestNested(t *testing.T) {
	c := NewConfig(&testConf{})
	t1, t2, t21, t3 := &testConf{}, &testConf{}, &testConf{}, &testConf{}
	sub1, sub2, sub21, sub3 := NewConfig(t1), NewConfig(t2), NewConfig(t21), NewConfig(t3)

	c.AddSub("sub1", sub1)
	c.AddSub("sub2", sub2)
	sub2.AddSub("sub2.1", sub21)
	c.AddSub("sub3", sub3)

	if err := json.Unmarshal([]byte(testJSON), c); err != nil {
		t.Fatal(err)
	}

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
