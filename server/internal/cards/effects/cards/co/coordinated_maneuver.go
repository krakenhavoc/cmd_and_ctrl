package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coordinated Maneuver — Instant {1}{W} (Tarkir: Dragonstorm):
//
//	"Choose one —
//	 • Coordinated Maneuver deals damage equal to the number of creatures
//	   you control to target creature or planeswalker.
//	 • Destroy target enchantment."
//
// The first mode counts the creatures its controller controls as the
// spell resolves (CR 608.2h), the target creature included; the second is
// a plain destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ff39311-565e-4414-b32b-eb24c70b3090",
		Name:         "Coordinated Maneuver",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.Mode("Coordinated Maneuver deals damage equal to the number of creatures you control to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
			effects.Mode("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				n := effects.CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsCreature() })
				return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: n}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
