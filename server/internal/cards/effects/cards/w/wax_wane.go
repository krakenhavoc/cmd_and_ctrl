package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wax // Wane — split card (Invasion, #5616):
//
//	Wax — Instant {G}: "Target creature gets +2/+2 until end of turn."
//	Wane — Instant {W}: "Destroy target enchantment."
//
// Each half is an ADR 0034 face (ADR 0103): Wax under the bare oracle
// ID, Wane under "#1". Only one half can be cast (CR 709.3; the card has
// no fuse). Wax is the shared boostTargetUntilEOT pump, Wane Tear's
// destroyChosenPermanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4744b086-fea2-4992-bfef-68c58cdf9cf0",
		Name:         "Wax",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BoostTargetUntilEOT("Wax — +2/+2", 2, 2)(ctx.Game, item)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "4744b086-fea2-4992-bfef-68c58cdf9cf0#1",
		Name:         "Wane",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}
