package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEchoesCopiesAColorlessSpellAndDoublesTheEntersTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Echoes of Eternity", OracleID: echoesOracle,
		TypeLine: "Kindred Enchantment — Eldrazi", ManaCost: "{3}{C}{C}{C}",
		Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 8; i++ {
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID})
	}
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Guidelight Matrix", "Artifact", guidelightOracle, nil)
	passPriorityAroundTable(t, g)
	matrices := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Guidelight Matrix" {
			matrices++
		}
	}
	if matrices != 2 {
		t.Errorf("%d Guidelight Matrix on the battlefield, want the spell and its token copy", matrices)
	}
	// Each enters trigger fires twice under Echoes: 2 permanents x 2 draws.
	if got := me.Hand.Size() - before; got != 4 {
		t.Errorf("hand grew by %d, want 4", got)
	}
}
