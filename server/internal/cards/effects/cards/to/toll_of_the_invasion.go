package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toll of the Invasion — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card.
//	 Amass Zombies 1."
//
// The revealed-hand pick (ADR 0116) filtered to nonland cards, then
// amass Zombies 1 (CR 701.47a) on the next line (ADR 0116 §6): the
// Army is yours and cannot change what may be chosen. You amass even
// when there is no nonland card to take (the 2019-05-03 ruling;
// CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a7c2e32-8585-4f7f-8635-199e3c6cd8a9",
		Name:         "Toll of the Invasion",
		Purpose:      game.Purpose{Amass: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		// The amass waits for the choice from the revealed hand (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ChooseNonlandFromTargetedHand),
			effects.Amass{Subtype: "Zombie", N: 1},
		),
	})
}
