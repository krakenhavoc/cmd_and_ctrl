package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harsh Scrutiny — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose a creature card
//	 from it. That player discards that card. Scry 1."
//
// The revealed-hand pick (ADR 0116) filtered to creature cards, then
// scry 1 (CR 701.22a) on the next line (ADR 0116 §6). The scry is a
// prompt of its own, queued behind the pick, so it is answered after
// the discard as printed (CR 608.2c). You scry even when there is no
// creature card to take (the 2016-09-20 ruling; CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31338e24-a437-4587-acab-cec46be021e7",
		Name:         "Harsh Scrutiny",
		Purpose:      game.Purpose{Scry: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		// The scry waits for the choice from the revealed hand (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ChooseFromRevealedHand{
					Player: effects.TargetedPlayer(ctx),
					Filter: effects.Creature(),
					Label:  "creature card",
				}.Apply(ctx)
			}),
			effects.Scry{N: 1},
		),
	})
}
