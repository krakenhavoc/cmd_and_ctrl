package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b17_helpers.go — helpers for the Standard set leftovers, batches 17
// and 18 (#4452). One block per seam; each names the cards that use it.

// selfWhileYourLifeAboveStarting is "as long as your life total is
// greater than your starting life total" (n = 1) or "is at least n
// greater than your starting life total" (n = 10): the condition for THIS
// creature alone (Elenda, Saint of Dusk). The starting total is the
// table's (ADR 0075), like Righteous Valkyrie's.
func selfWhileYourLifeAboveStarting(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && controllerLifeAtLeast(g, source, startingLifeOf(g)+n)
	}
}

// stdDifferentManaValuesAmongYourNonlands is "the number of different
// mana values among nonland permanents you control" (Lunar Insight): the
// distinct mana values read through the layers, an {X} counted as 0 and a
// token with no mana cost as 0 (CR 202.3).
func stdDifferentManaValuesAmongYourNonlands(g *game.Game, you uuid.UUID) int {
	seen := map[int]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you || c.IsLand() {
			continue
		}
		mv, _ := g.ManaValueForEffect(c)
		seen[mv] = true
	}
	return len(seen)
}

// CountersIfCast is "this creature enters with <n> <kind> counters on it
// if you cast it" (Nine-Lives Familiar): a CR 614.1c clause read off the
// resolving spell's cast record, from any zone. A permanent that was not
// cast (reanimated, blinked, a token, a CR 707.10 copy) enters with none.
func CountersIfCast(kind string, n int) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind: kind,
		Count: func(cast game.CastCounts) int {
			if cast.CastFrom == "" {
				return 0
			}
			return n
		},
	}
}

// stdYouControlAPower4Creature is the trigger-time half of "if you control
// a creature with power 4 or greater" (Flamewake Phoenix's ferocious):
// current power, counters and anthems included. The resolution-time half
// is CondPowerAtLeast.
func stdYouControlAPower4Creature(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
}
