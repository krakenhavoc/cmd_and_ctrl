package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// destroyed_this_way_test.go — #815 at the catalog level.
//
// The engine-side cases live in game/destroyed_this_way_test.go. This
// is the one a player would file the bug from: a real Fumigate,
// resolved through the real priority loop, over a board where the
// CR 614 window takes two of the creatures away from the destruction
// in the two ways it can — cancelling it outright, and sending the
// permanent somewhere that is not a graveyard.
//
// Indestructible is the case that was already covered
// (s30_indestructible_wipes_test.go) and it is a different mechanism:
// it is filtered out BEFORE the sweep opens. What is pinned here is a
// creature the window saves once the sweep is already running, which
// no filter can see coming.

// registerExitReplacement installs a battlefield-exit replacement
// gated to one card for the duration of one test.
func registerExitReplacement(t *testing.T, g *game.Game, only uuid.UUID, label string, rewrite func(ev *game.ReplacementEvent)) {
	t.Helper()
	registerExitReplacementFrom(t, g, game.ZoneBattlefield, only, label, rewrite)
}

// registerExitReplacementFrom is the same helper with the source zone
// named: #911's exiles leave a GRAVEYARD, which is the Rest in Peace /
// Leyline family's zone rather than the battlefield.
func registerExitReplacementFrom(t *testing.T, g *game.Game, from game.ZoneKind, only uuid.UUID, label string, rewrite func(ev *game.ReplacementEvent)) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove &&
					ev.CardID == only &&
					ev.OldZone == from
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				rewrite(ev)
				return nil
			},
			Label: label,
		})
	})
}

// TestFumigateDoesNotPayForADestructionTheWindowTookAway is the
// issue's own test, on the card the clause is printed on. Three
// creatures are swept: one is destroyed, one has its destruction
// cancelled mid-window ("it can't be destroyed"), one is exiled
// instead. Fumigate gains 1 life, not 3.
//
// The exiled one is the judgement call, and it is CR 701.7a's: to
// destroy a permanent is to move it from the battlefield to its
// owner's GRAVEYARD, so a destruction a replacement rewrote into an
// exile destroyed nothing however thoroughly the permanent left.
func TestFumigateDoesNotPayForADestructionTheWindowTookAway(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doomed := pushWipeCreature(g, me.ID, "Doomed", "Creature — Bear", 2, 2)
	saved := pushWipeCreature(g, opp.ID, "Saved", "Creature — Bear", 2, 2)
	exiled := pushWipeCreature(g, opp.ID, "Exiled", "Creature — Bear", 2, 2)

	registerExitReplacement(t, g, saved, "it can't be destroyed",
		func(ev *game.ReplacementEvent) { ev.Cancel() })
	registerExitReplacement(t, g, exiled, "exile it instead",
		func(ev *game.ReplacementEvent) {
			ev.NewZone = game.ZoneExile
			ev.NewZoneOwner = uuid.Nil
		})

	before := me.Life
	castCatalogSpell(t, g, "Fumigate", "Sorcery", fumigateOracle, nil)
	passPriorityAroundTable(t, g)

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
			"the saved one never left and the exiled one was never put into a graveyard (CR 701.7a)",
			before, me.Life, want)
	}
}
