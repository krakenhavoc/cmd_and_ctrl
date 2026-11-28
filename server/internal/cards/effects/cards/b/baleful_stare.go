package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baleful Stare — Sorcery {2}{U} (Portal, S82 OLD batch 783, #5630):
//
//	"Target opponent reveals their hand. You draw a card for each Mountain and
//	 red card in it."
//
// The hand is revealed to the table (CR 701.20a) and counted once: a card that
// is a Mountain card or a red card counts once, and a card that is both (a red
// card with the Mountain type) still counts once. The caster draws that many.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4135131d-4653-4767-ad1d-68c9bf393c3a",
		Name:         "Baleful Stare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB783RevealHandDrawPerMatching(ctx, effects.S82OldB783LandOrColor("Mountain", "R"))
		},
	})
}
