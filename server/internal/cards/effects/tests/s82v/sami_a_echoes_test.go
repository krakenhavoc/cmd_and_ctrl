package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEchoesCopiesAColorlessSpellAndDoublesTheEntersTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Echoes of Eternity", OracleID: effectstest.EchoesOracle,
		TypeLine: "Kindred Enchantment — Eldrazi", ManaCost: "{3}{C}{C}{C}",
		Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID})
	}
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Guidelight Matrix", "Artifact", effectstest.GuidelightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
