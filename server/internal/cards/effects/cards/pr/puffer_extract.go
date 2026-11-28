package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Puffer Extract — Artifact {5} (Mercadian Masques, S82 OLD b681, #5619):
//
//	"{X}, {T}: Target creature you control gets +X/+X until end of turn.
//	 Destroy it at the beginning of the next end step."
//
// X is announced with the cost (CR 107.3k). The pump is layer 7c until end of
// turn; the destruction is Stone Giant's delayed trigger, naming the creature
// as it is now (CR 603.7c), so a creature that has left and returned is not
// destroyed. A target that is illegal on resolution gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f427de6-b551-4699-b78a-44ff668198d8",
		Name:         "Puffer Extract",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Target creature you control gets +X/+X until end of turn. Destroy it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.PumpXUntilEOT(0))},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.OldB681PumpThenDestroyAtEndStep(ctx, "Puffer Extract")
			},
		}},
	})
}
