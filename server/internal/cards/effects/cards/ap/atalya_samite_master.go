package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atalya, Samite Master — Legendary Creature — Human Cleric {3}{W}{W}, 2/3
// (Invasion, #5616):
//
//	"{X}, {T}: Choose one —
//	 • Prevent the next X damage that would be dealt to target creature this
//	   turn. Spend only white mana on X.
//	 • You gain X life. Spend only white mana on X."
//
// Crypt Rats' restriction on {X} (SpendOnlyOnX), so only white mana can pay for
// X; the mode is chosen on announce and only the prevention names a target.
// The prevention is Battlefield Medic's shield with X read from the ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84587724-af96-4396-a0b3-1133d492d915",
		Name:         "Atalya, Samite Master",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Choose one — Prevent the next X damage that would be dealt to target creature this turn; or you gain X life. Spend only white mana on X.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost(), effects.SpendOnlyOnX("W")),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Prevent the next X damage that would be dealt to target creature this turn.",
					effects.TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
							return nil
						}
						return effects.PreventNextDamage{Target: t.ID, Amount: ctx.X(), Label: "Atalya, Samite Master: prevent the next X damage"}.Apply(ctx)
					}),
				effects.ModeDoing("You gain X life.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.GainLife{Amount: ctx.X()}.Apply(ctx)
					}),
			),
		}},
	})
}
