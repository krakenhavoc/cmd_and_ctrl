package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b20_more_test.go — the rest of batch 20 (#5479): Jace, Kaya, Lukka,
// Nahiri, Nissa, Goliath Hatchery and Viral Spawning.

// s82CastWalker casts a planeswalker from hand for `cost`, strictly, with
// `life` Phyrexian symbols paid with life, and resolves it.
func S82CastWalker(t *testing.T, g *game.Game, name, oracle, typeLine, cost string, loyalty, life int, pool string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, StartingLoyalty: loyalty,
		OracleID: oracle, Owner: me.ID, Controller: me.ID,
	})
	ToMainForCost(t, g)
	FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, PhyrexianLife: life}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}
