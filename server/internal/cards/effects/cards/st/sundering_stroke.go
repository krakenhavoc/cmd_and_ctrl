package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sundering Stroke — Sorcery {6}{R} (Throne of Eldraine, #5516):
//
//	"Sundering Stroke deals 7 damage divided as you choose among one, two,
//	 or three targets. If at least seven red mana was spent to cast this
//	 spell, instead Sundering Stroke deals 7 damage to each of those
//	 permanents and/or players."
//
// The division is announced either way (CR 601.2d). With seven red mana
// spent (Slaying Fire's adamant read) every legal target takes the full 7.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "9c348bc7-e6e1-40a2-9798-6d2c4a2d8e72",
		Name:           "Sundering Stroke",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"R"},
		Targets:        effects.TargetAny().WithCount(1, 3).Dividing(effects.Divide(7)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if !effects.AdamantSpent(ctx, "R", 7) {
				return effects.DealDividedDamage(ctx)
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, t := range ctx.LegalTargets() {
					if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 7}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
