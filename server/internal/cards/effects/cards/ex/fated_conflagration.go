package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fated Conflagration — Instant {1}{R}{R}{R} (Born of the Gods,
// #5559):
//
//	"Fated Conflagration deals 5 damage to target creature or
//	 planeswalker. If it's your turn, scry 2."
//
// Jaya's Firenado with the scry behind "if it's your turn", read as
// the spell resolves (IsYourTurn, Feed the Cauldron's check).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.DamageToTarget(0, 5))
	purpose.Scry = 2
	effects.Register(effects.Spec{
		OracleID:     "ac0a110a-0721-4aa6-a891-95e3fcc847df",
		Name:         "Fated Conflagration",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			if !effects.IsYourTurn(ctx.Game, item.Controller) {
				return nil
			}
			return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
