package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanctuary Lockdown — Enchantment {2}{W} (Ikoria: Lair of Behemoths,
// #5512):
//
//	"Humans you control get +1/+1.
//	 {2}, Tap two untapped Humans you control: Tap target creature an
//	 opponent controls."
//
// A layer 7c anthem over the Humans its controller controls (the lord
// shape, TribalAnthem) and an activated ability whose cost is Whirler
// Rogue's TapOthers: two untapped Humans, summoning sickness not applying
// because it is not a {T} cost (CR 302.6, 702.62).
//
// No simplification.
func init() {
	const tapTwo = "two untapped Humans you control"
	effects.Register(effects.Spec{
		OracleID:     "66793ddf-aca3-44cb-8742-4eee178f31c6",
		Name:         "Sanctuary Lockdown",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Human"}, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, Tap two untapped Humans you control: Tap target creature an opponent controls.",
			Cost: effects.Plus(effects.ManaCost("{2}"), game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent(tapTwo, effects.HasSubtype("Human")),
				Label:  tapTwo,
			}}),
			Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
