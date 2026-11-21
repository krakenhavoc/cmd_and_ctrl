package testsds

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The token is the TARGETED player's, not the caster's — the
// political half of the card and the part a two-clause statement
// exists to get right.
func TestEchocastingSymposiumGivesTheTokenToTheTargetedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Big Wurm", "Creature — Wurm", "", 6, 6)

	effectstest.CastCatalogSpell(t, g, "Echocasting Symposium", "Sorcery — Lesson", effectstest.EchocastingSymposiumOracle,
		[]game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
			{Kind: game.TargetCard, ID: mine, Slot: 1},
		})
	effectstest.PassPriorityAroundTable(t, g)

	var tokens int
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Big Wurm" && effects.IsToken(c) {
			tokens++
			if c.Controller != opp.ID {
				t.Errorf("the copy is the targeted player's, not the caster's (%v)", c.Controller)
			}
			if c.Power != 6 || c.Toughness != 6 {
				t.Errorf("the copy is %d/%d, want 6/6", c.Power, c.Toughness)
			}
		}
	}
	if tokens != 1 {
		t.Errorf("created %d copies, want 1", tokens)
	}
}

// The creature clause is "you control": an opponent's creature is not
// a legal pick, refused at announce (CR 601.2c).
func TestEchocastingSymposiumOnlyCopiesACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)

	err := effectstest.CastCatalogSpellErr(t, g, "Echocasting Symposium", "Sorcery — Lesson", effectstest.EchocastingSymposiumOracle,
		[]game.TargetRef{
			{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
			{Kind: game.TargetCard, ID: theirs, Slot: 1},
		})
	if err == nil {
		t.Fatal("copying a creature you do not control is refused at announce")
	}
}
