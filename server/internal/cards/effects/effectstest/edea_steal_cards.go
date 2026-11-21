package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// edea_steal_cards_test.go pins the Edea deck's steal-and-copy slice
// (E2, #1565).

// e2Zone reports the kind of zone a card is in, or "" when it is
// nowhere.
func E2Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// e2Destroy destroys a permanent and settles whatever that triggered.
func E2Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	// The CR 704.3 boundary a destroying resolution ends at, where an
	// "until" return is owed (CR 610.3, #1729) whether or not anything
	// reached the stack.
	g.RunStateChecksForTest()
	PassPriorityAroundTable(t, g)
}
