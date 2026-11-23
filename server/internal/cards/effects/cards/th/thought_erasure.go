package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thought Erasure — Sorcery {U}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card.
//	 Surveil 1."
//
// The revealed-hand pick (ADR 0116) filtered to nonland cards, then
// surveil 1 (CR 701.25a) on the next line (ADR 0116 §6). The surveil
// is a prompt of its own, queued behind the pick, so it is answered
// after the discard as printed (CR 608.2c). You surveil even when
// nothing is discarded (the 2018-10-05 ruling; CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4bae4e34-fcf4-4aee-8da2-5b3ee41a595a",
		Name:         "Thought Erasure",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		// The surveil waits for the choice from the revealed hand (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ChooseNonlandFromTargetedHand),
			effects.Surveil{N: 1},
		),
	})
}
