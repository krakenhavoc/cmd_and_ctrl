package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinning Wheel — Artifact {3} (Throne of Eldraine, #5516):
//
//	"{T}: Add one mana of any color.
//	 {5}, {T}: Tap target creature."
//
// A Mana Geode's mana ability and the shared tap row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0935a0f-7087-40cd-a094-89f200da6ef0",
		Name:         "Spinning Wheel",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
