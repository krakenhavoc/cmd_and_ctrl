package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seed Spark — Instant {3}{W} (Ravnica: City of Guilds, S82 OLD b582):
//
//	"Destroy target artifact or enchantment. If {G} was spent to cast this
//	 spell, create two 1/1 green Saproling creature tokens."
//
// Rip Apart's destroy through the single-target verb, then River's Grasp's
// colour reading (ReadsOwnPayment; WantsColors so an auto-paid cast reaches
// for a green source). The tokens are created whether or not the target was
// still there to destroy, as the second sentence is not conditional on the
// first; with the one target illegal the whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "fc2e441c-bc0b-439f-af71-daba3b033381",
		Name:           "Seed Spark",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"G"},
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Targets:        effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			if ctx.ManaSpentOfColor("G") == 0 {
				return nil
			}
			return effects.CreateToken{Controller: item.Controller, Template: effects.B11GreenSaprolingToken(), N: 2}.Apply(ctx)
		},
	})
}
