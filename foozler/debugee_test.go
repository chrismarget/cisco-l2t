package foozler

import (
	"testing"
)

func TestWhatever(t *testing.T) {
	s := removeTimestamp("Aug 21 15:45:09.385: trace_request->src_mac     : ffff.ffff.ffff")

	exp := "trace_request->src_mac     : ffff.ffff.ffff"
	if s != exp {
		t.Fatalf("expected '%s', got '%s'", exp, s)
	}
}
