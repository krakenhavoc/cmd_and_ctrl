package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pr8cPermanent puts a catalogued creature with printed `power` /
// `toughness` and `counters` +1/+1 counters on the battlefield for
// `owner`, as if it had entered (no entry replacement runs).
func Pr8cPermanent(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, power, toughness, counters int) uuid.UUID {
	t.Helper()
	id := ApaPush(g, owner, owner, game.Card{Name: name, OracleID: oracle, TypeLine: "Creature — Test", Power: power, Toughness: toughness})
	if counters > 0 {
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(id, game.CounterPlusOne, counters); err != nil {
				t.Fatal(err)
			}
		})
	}
	return id
}

// pr8cThree is three opposing creatures, three sources of damage.
func Pr8cThree(g *game.Game, opp uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	return Pr7Creature(g, opp, "A", 2, "R"), Pr7Creature(g, opp, "B", 2, "G"), Pr7Creature(g, opp, "C", 2, "W")
}
