package main

import "testing"

func TestVersionLess(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{{"2.1.0", "v2.2.0", true}, {"2.1.0", "v2.1.0", false}, {"2.10.0", "v2.9.9", false}, {"2.1.0", "v3", true}} {
		if got := versionLess(c.a, c.b); got != c.want {
			t.Errorf("versionLess(%s,%s)=%v", c.a, c.b, got)
		}
	}
}
