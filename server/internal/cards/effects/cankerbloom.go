package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cankerbloom — Creature — Phyrexian Fungus {1}{G}, 3/2:
//
//	"{1}, Sacrifice this creature: Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Proliferate."
//
// A modal activated ability (ADR 0065): the mode and its target are
// announced with the activation, and the sacrifice and the {1} are
// paid up front, so the Cankerbloom is gone whatever the answer.
// The two destroy bullets are Bant Charm's shared body; the
// proliferate bullet has no target.
//
// The proliferate bullet asks the player what to proliferate (#2525).
func init() {
	Register(Spec{
		OracleID:     "d5b80895-621a-40df-bf48-6c7295658f21",
		Name:         "Cankerbloom",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, Sacrifice this creature: Choose one — destroy target artifact; destroy target enchantment; or proliferate.",
			Cost:  Plus(ManaCost("{1}"), SacrificeThis()),
			Modes: ChooseOne(
				ModeDoing("Destroy target artifact.",
					TargetPermanent("target artifact", Artifact()),
					DestroyTheModesTarget),
				ModeDoing("Destroy target enchantment.",
					TargetPermanent("target enchantment", Enchantment()),
					DestroyTheModesTarget),
				ModeDoing("Proliferate.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error {
						return Proliferate{}.Apply(ctx)
					}),
			),
		}},
	})
}
