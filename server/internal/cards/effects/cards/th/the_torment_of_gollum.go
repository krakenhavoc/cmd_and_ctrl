package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Torment of Gollum — Sorcery {3}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card.
//	 Amass Orcs 2."
//
// The revealed-hand pick (ADR 0116) filtered to nonland cards, then
// amass Orcs 2 (CR 701.47a) on the next line (ADR 0116 §6): the Army is
// yours and cannot change what may be chosen. You amass even when
// nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b70aad1-bcb1-4359-b27e-ba8b40ef2752",
		Name:         "The Torment of Gollum",
		Purpose:      game.Purpose{Amass: 2},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		// The Army waits for the question the choice asks (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ChooseNonlandFromTargetedHand(ctx)
			}),
			effects.Amass{Subtype: "Orc", N: 2},
		),
	})
}
