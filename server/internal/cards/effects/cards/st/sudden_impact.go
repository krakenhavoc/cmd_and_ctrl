package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sudden Impact — Instant {3}{R} (Eighth Edition, S82 OLD batch 610, #5607):
//
//	"Sudden Impact deals damage to target player equal to the number of cards
//	 in that player's hand."
//
// Gaze of Adamaro's and Vicious Shadows' body (OldB515DamageEqualToTargetsHand):
// the hand is counted as the spell resolves (CR 608.2b), and a target player
// who has left the game takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2bf65e7d-6d31-4ad6-93cd-623612d3175d",
		Name:         "Sudden Impact",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB515DamageEqualToTargetsHand(ctx.Game, item)
		},
	})
}
