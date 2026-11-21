package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// selfless_squire_test.go — CR 615.13 (#2047): Selfless Squire's
// "Whenever damage that would be dealt to you is prevented, put that many
// +1/+1 counters on this creature."

const SelflessSquireOracle = "7d079c78-b2a9-444f-ab54-c61c041baa6d"

// squireTriggers counts the Squire's triggered items waiting or on the
// stack.
func SquireTriggers(g *game.Game, squire uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.SourceCardID == squire {
				n++
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered && it.SourceCardID == squire {
				n++
			}
		}
	})
	return n
}

// castSquire casts the Squire for seat 0 and resolves it and its enters
// trigger: no damage to that player this turn.
func CastSquire(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, "Selfless Squire", "Creature — Human Soldier", SelflessSquireOracle, nil)
	PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, id) {
		t.Fatal("the Squire did not resolve")
	}
	return id
}
