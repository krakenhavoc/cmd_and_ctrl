package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Olag and Miau, New Friends — Legendary Creature — Zombie Cat {W}{B}{G},
// 3/3 (Mystery Booster Commander Edition):
//
//	"Whenever another Zombie you control enters, put a +1/+1 counter on
//	 target Cat you control.
//	 Whenever another Cat you control enters, put a +1/+1 counter on target
//	 Zombie you control.
//	 At the beginning of your end step, each opponent loses life equal to
//	 the number of tapped Cats and/or Zombies you control."
//
// Two triggers, one per printed sentence; a creature that is both a Zombie
// and a Cat fires both. Olag and Miau is both, so it can be its own
// target. The end-step count is a creature that is a Cat, a Zombie or both
// counted once, as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d117825e-fb7d-44dc-a17d-71bd610a6078",
		Name:         "Olag and Miau, New Friends",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			olagAndMiauTrigger("Zombie", "Cat"),
			olagAndMiauTrigger("Cat", "Zombie"),
			AtYourEndStep("Olag and Miau, New Friends — each opponent loses life equal to the number of tapped Cats and/or Zombies you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					cat, zombie := ctx.TextSubtype("Cat"), ctx.TextSubtype("Zombie")
					n := countControlled(g, item.Controller, func(c game.Card) bool {
						return c.Tapped && (c.HasSubtype(cat) || c.HasSubtype(zombie))
					})
					return EachOpponentLosesLife{N: n}.Apply(ctx)
				}),
		},
	})
}

// olagAndMiauTrigger is "whenever another <entering> you control enters,
// put a +1/+1 counter on target <target> you control."
func olagAndMiauTrigger(entering, target string) game.TriggeredAbility {
	label := "Olag and Miau, New Friends — put a +1/+1 counter on target " + target + " you control"
	return Targeting(
		On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype(entering))
		}, label, onePlusOneCounterOnFirstLegalTarget),
		TargetCreature("target "+target+" you control", YouControl(), OfSubtype(target)))
}
