package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ingenious Leonin — Creature — Cat Soldier {4}{W}, 4/4 (Foundations,
// #4465):
//
//	"{3}{W}: Put a +1/+1 counter on another target attacking creature you
//	 control. If that creature is a Cat, it gains first strike until end
//	 of turn."
//
// The target is another attacking creature you control (Another, CR
// 109.1), checked as it is chosen and again on resolution (CR 608.2b).
// The counter is Vastwood Fortification's (plusOneCounterOnTheTarget)
// and goes on first; whether the creature is a Cat is read after
// it, through the layers, so a changeling counts. "Cat" is a word of the
// card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99853d2e-2223-4c05-bbd7-0a746f8f810d",
		Name:         "Ingenious Leonin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Put a +1/+1 counter on another target attacking creature you control. If that creature is a Cat, it gains first strike until end of turn.",
			Cost:    effects.ManaCost("{3}{W}"),
			Targets: effects.Another(effects.TargetCreature("another target attacking creature you control", effects.AttackingCreature(), effects.YouControl())),
			Effect: effects.Do(
				effects.Instruction(effects.PutPlusOneCounterOnTheTarget),
				// The Cat test waits for a prompt the counter asks (CR 608.2c).
				effects.Instruction(func(ctx *effects.Context) error {
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok || !effects.OnBattlefield(ctx.Game, id) {
						return nil
					}
					ctx.Game.RecomputeLayersIfStaleLocked()
					c, ok := ctx.Game.LookupCardForEffect(id)
					if !ok || !c.HasSubtype(ctx.TextSubtype("Cat")) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   id,
						Keywords: []string{"first strike"},
						Label:    "Ingenious Leonin — first strike until end of turn",
					}.Apply(ctx)
				}),
			),
		}},
	})
}
