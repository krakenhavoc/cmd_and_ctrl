package effectstest

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
func RegisterExitReplacement(t *testing.T, g *game.Game, only uuid.UUID, label string, rewrite func(ev *game.ReplacementEvent)) {
	t.Helper()
	RegisterExitReplacementFrom(t, g, game.ZoneBattlefield, only, label, rewrite)
}

// registerExitReplacementFrom is the same helper with the source zone
// named: #911's exiles leave a GRAVEYARD, which is the Rest in Peace /
// Leyline family's zone rather than the battlefield.
func RegisterExitReplacementFrom(t *testing.T, g *game.Game, from game.ZoneKind, only uuid.UUID, label string, rewrite func(ev *game.ReplacementEvent)) {
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
