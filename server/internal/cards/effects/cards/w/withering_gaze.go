package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Withering Gaze — Sorcery {2}{U} (Portal, S82 OLD batch 783, #5630):
//
//	"Target opponent reveals their hand. You draw a card for each Forest and
//	 green card in it."
//
// Baleful Stare's count for Forest and green: the hand is revealed to the table
// (CR 701.20a), each card that is a Forest card or a green card counts once, and
// the caster draws that many.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9225016d-adfd-43c6-99cd-d41a7e0d35d6",
		Name:         "Withering Gaze",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB783RevealHandDrawPerMatching(ctx, effects.S82OldB783LandOrColor("Forest", "G"))
		},
	})
}
