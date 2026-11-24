package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Memory Lapse — Instant {1}{U}:
//
//	"Counter target spell. If that spell is countered this way, put it on top of
//	 its owner's library instead of into that player's graveyard."
//
// Hinder's counter-to-library with the one position the card prints, the top
// (game.LibraryPlaceTop), so there is no choice to ask. The spell is still on the
// stack while its fate is decided: a commander's owner is offered the command
// zone (CR 903.9), flashback's exile wins, and a spell that can't be countered
// is neither countered nor moved. The card goes to its OWNER's library, not the
// caster's. A copy of a spell is countered and ceases to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbfb3e4a-b389-4391-8141-13b68c0ef2e0",
		Name:         "Memory Lapse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CounterToLibrary{
				StackID:   item.Targets[0].ID,
				Placement: game.LibraryPlaceTop,
				Label:     "Memory Lapse — put it on top of its owner's library",
			}.Apply(ctx)
		},
	})
}
