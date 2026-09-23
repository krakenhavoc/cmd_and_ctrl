package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_instead_of_leaving.go — #1221: "if it would leave the
// battlefield, exile it instead of putting it anywhere else", pinned
// to ONE permanent.
//
// Two cards print the clause and both got it from a reanimation that
// is meant to be temporary: Whip of Erebos (#296) and every unearth
// card (CR 702.82a). It was written inline on the first; the second
// would have been a second copy of a clause whose failure mode is
// silent — an omitted `NewZone != ZoneExile` guard is an infinite
// loop, and a missed event kind is a creature that can be sacrificed
// back into a second reanimation. One helper, so the two cannot
// drift.
//
// TURN-SCOPED, which is a real (declared) limitation rather than a
// choice: the registry is swept at cleanup (Game.ClearTurnScopedReplacementsLocked),
// and the redirect covers the whole window the printed clause cares
// about because the end-step exile fires before that sweep. It lapses
// only if the permanent somehow survives the turn, which no catalog
// card can arrange today.
//
// A BOUNCE bypasses it, and that is an engine seam rather than a card
// decision: BounceToHandForEffect moves the card without running the
// CR 614 pipeline, so a whipped or unearthed creature returned to
// hand in response goes to hand, not exile. Whip of Erebos declared
// it first; every card built on this helper inherits the caveat.

// ExileInsteadOfLeavingBattlefield registers the CR 614 replacement
// that redirects ANY battlefield exit of the one instance `cardID` to
// exile, under `controller`'s control of the replacement effect.
//
// The instance ID is captured, never a *Card: undo resolves the
// closures against a cloned game, where an instance ID is stable and
// a pointer is not.
//
// `NewZone != ZoneExile` is the guard that stops the effect replacing
// its own result, which CR 614.5 would forbid anyway and which would
// be an infinite loop if it did not.
//
// Caller must hold g.mu — every caller is an ability body running at
// resolution, which does.
func ExileInsteadOfLeavingBattlefield(g *game.Game, cardID, controller uuid.UUID, label string) {
	g.RegisterTurnScopedReplacement(game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventMove && ev.CardID == cardID &&
				ev.OldZone == game.ZoneBattlefield && ev.NewZone != game.ZoneExile
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneExile
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			return controller
		},
		Label: label,
	})
}
