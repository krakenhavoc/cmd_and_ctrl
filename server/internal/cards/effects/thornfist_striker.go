package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thornfist Striker — Creature — Elf Druid {2}{G}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"Ward {1}
//	 Infusion — Creatures you control get +1/+0 and have trample as long as
//	 you gained life this turn."
//
// Ward is the shared Ward trigger. Infusion is an ability word (CR 207.2c):
// two continuous effects of the same static ability, +1/+0 (layer 7c) and
// trample (layer 6), both applying to every creature you control, the Striker
// included, while you have gained life this turn. They are live statics: they
// turn on the moment the first point of life is gained (the layer cache is
// told about life changes, DependsOnLifeTotal) and off when the turn ends,
// and a creature that enters afterwards is covered too (CR 611.3).
//
// No simplification.
func init() {
	applies := socB38InfusionAppliesToCreaturesYouControl
	Register(Spec{
		OracleID:     "d0d13d5d-677d-4e1a-9e58-2f9840581f6d",
		Name:         "Thornfist Striker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Thornfist Striker — ward {1}"),
		},
		Static: []game.StaticAbility{
			socB38InfusionStatic(b16Anthem(applies, 1, 0), "Infusion — creatures you control get +1/+0 as long as you gained life this turn."),
			socB38InfusionStatic(b16GrantKeywords(applies, "trample"), "Infusion — creatures you control have trample as long as you gained life this turn."),
		},
	})
}
