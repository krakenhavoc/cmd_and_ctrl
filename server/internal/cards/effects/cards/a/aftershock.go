package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aftershock — Sorcery {2}{R}{R} (Jumpstart 2022, #5948):
//
//	"Destroy target artifact, creature, or land. Aftershock deals 3 damage
//	 to you."
//
// The destroy is Banishment Decree's target clause with a land; the 3
// damage to the controller is dealt by the spell, after the destroy, and
// only when the spell resolves (a fizzled Aftershock hurts nobody).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6bef0c4-3dfc-49a8-8fec-22b2570b704b",
		Name:         "Aftershock",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Controller, Amount: 3}.Apply(ctx)
		},
	})
}
