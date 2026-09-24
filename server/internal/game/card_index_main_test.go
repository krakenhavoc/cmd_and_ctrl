package game

import (
	"os"
	"testing"
)

// TestMain runs every test in this package with the card index's
// cross-check mode on (#1479, ADR 0094): each card lookup also runs the
// zone walk the index replaced, and a disagreement panics at the
// lookup that caused it. The index is a hint that is checked on every
// read, so this can only fire if an instance ID is in two zones at
// once or the index itself is wrong — either way, a bug worth the
// stack trace.
func TestMain(m *testing.M) {
	SetCardIndexCrossCheck(func(msg string) { panic(msg) })
	os.Exit(m.Run())
}
