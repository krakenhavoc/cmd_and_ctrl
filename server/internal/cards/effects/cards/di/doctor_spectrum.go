package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doctor Spectrum — Legendary Creature — Human Hero {3}{W}{W}, 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 When Doctor Spectrum enters, choose one —
//	 • Create a 0/4 colorless Wall creature token with defender.
//	 • Put a +1/+1 counter on each other Hero you control.
//	 • Destroy target enchantment."
//
// A modal ETB (CR 603.3c). The counter bullet is untargeted and reads
// "Hero" through the source's text and each creature's effective
// subtypes, so a changeling counts. Only the third bullet targets, so
// with no enchantment on the battlefield the first two are still offered.
//
// No simplification.
func init() {
	spectrum := effects.WhenThisEnters("Doctor Spectrum — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	spectrum.Modes = effects.ChooseOne(
		effects.ModeDoing("Create a 0/4 colorless Wall creature token with defender.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Template: effects.TokenCard("0/4 colorless Wall with defender"), N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Put a +1/+1 counter on each other Hero you control.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				hero := effects.And(effects.Creature(), effects.HasSubtype(ctx.TextSubtype("Hero")))
				return effects.MshB1CounterOnEachOtherYouControl(hero)(ctx.Game, item)
			}),
		effects.ModeDoing("Destroy target enchantment.",
			effects.TargetPermanent("target enchantment", effects.Enchantment()),
			effects.DestroyTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:        "fad80e75-8d08-4a0b-a05e-14f4c96947bf",
		Name:            "Doctor Spectrum",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{spectrum},
	})
}
