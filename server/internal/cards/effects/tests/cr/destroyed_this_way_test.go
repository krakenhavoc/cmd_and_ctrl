package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestFumigateDoesNotPayForADestructionTheWindowTookAway is the
// issue's own test, on the card the clause is printed on. Three
// creatures are swept: one is destroyed, one has its destruction
// cancelled mid-window ("it can't be destroyed"), one is exiled
// instead. Fumigate gains 1 life, not 3.
//
// The exiled one is the judgement call, and it is CR 701.8a's: to
// destroy a permanent is to move it from the battlefield to its
// owner's GRAVEYARD, so a destruction a replacement rewrote into an
// exile destroyed nothing however thoroughly the permanent left.
func TestFumigateDoesNotPayForADestructionTheWindowTookAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doomed := effectstest.PushWipeCreature(g, me.ID, "Doomed", "Creature — Bear", 2, 2)
	saved := effectstest.PushWipeCreature(g, opp.ID, "Saved", "Creature — Bear", 2, 2)
	exiled := effectstest.PushWipeCreature(g, opp.ID, "Exiled", "Creature — Bear", 2, 2)

	effectstest.RegisterExitReplacement(t, g, saved, "it can't be destroyed",
		func(ev *game.ReplacementEvent) { ev.Cancel() })
	effectstest.RegisterExitReplacement(t, g, exiled, "exile it instead",
		func(ev *game.ReplacementEvent) {
			ev.NewZone = game.ZoneExile
			ev.NewZoneOwner = uuid.Nil
		})

	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Fumigate", "Sorcery", effectstest.FumigateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(doomed) {
		t.Error("the unprotected creature is destroyed — control case broken")
	}
	if !g.Battlefield.Contains(saved) {
		t.Error("a destruction the window cancelled leaves its permanent on the battlefield")
	}
	if !g.Exile.Contains(exiled) {
		t.Error("the redirected destruction put its permanent in exile")
	}
	if want := before + 1; me.Life != want {
		t.Errorf("life %d -> %d, want %d — one creature was destroyed this way; "+
			"the saved one never left and the exiled one was never put into a graveyard (CR 701.8a)",
			before, me.Life, want)
	}
}
