package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lapse of Certainty — Instant {2}{W} (Conflux, #5587):
//
//	"Counter target spell. If that spell is countered this way, put it on
//	 top of its owner's library instead of into that player's
//	 graveyard."
//
// Spell Crumple's CounterToLibrary at the top. A spell that can't be
// countered is not moved; the countered card goes through the shared
// stack exit, so a commander's owner may still choose the command zone
// (CR 903.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b721ad3-d0f6-4eec-9bf0-f57ea9dfa392",
		Name:         "Lapse of Certainty",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CounterToLibrary{
				StackID:   item.Targets[0].ID,
				Placement: game.LibraryPlaceTop,
				Label:     "Lapse of Certainty — put it on top of its owner's library",
			}.Apply(ctx)
		},
	})
}
