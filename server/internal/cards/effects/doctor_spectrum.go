package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	spectrum := WhenThisEnters("Doctor Spectrum — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	spectrum.Modes = ChooseOne(
		ModeDoing("Create a 0/4 colorless Wall creature token with defender.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Template: TokenCard("0/4 colorless Wall with defender"), N: 1}.Apply(ctx)
			}),
		ModeDoing("Put a +1/+1 counter on each other Hero you control.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				hero := And(Creature(), HasSubtype(ctx.TextSubtype("Hero")))
				return mshB1CounterOnEachOtherYouControl(hero)(ctx.Game, item)
			}),
		ModeDoing("Destroy target enchantment.",
			TargetPermanent("target enchantment", Enchantment()),
			DestroyTheModesTarget),
	)
	Register(Spec{
		OracleID:        "fad80e75-8d08-4a0b-a05e-14f4c96947bf",
		Name:            "Doctor Spectrum",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{spectrum},
	})
}
