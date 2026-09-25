package effects

import (
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestMain runs every test in this package with the card index's
// cross-check mode on (#1479, ADR 0094): each card lookup also runs the
// zone walk the index replaced, and a disagreement panics at the
// lookup that caused it. The index is a hint that is checked on every
// read, so this can only fire if an instance ID is in two zones at
// once or the index itself is wrong — either way, a bug worth the
// stack trace.
func TestMain(m *testing.M) {
	game.SetCardIndexCrossCheck(func(msg string) { panic(msg) })
	// ADR 0041 P9 (tier 4-2): the legacy-trigger lint reads the
	// PRODUCTION catalog, so it is captured before any test registers
	// a fixture of its own.
	snapshotProductionDefKeys()
	os.Exit(m.Run())
}
