package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drover of the Mighty — Creature — Human Druid {1}{G}, 1/1 (Lost Caverns
// of Ixalan Commander):
//
//	"This creature gets +2/+2 as long as you control a Dinosaur.
//	 {T}: Add one mana of any color."
//
// A layer 7c self pump (CR 613.4c) while its controller controls a
// Dinosaur; the condition reads the layered view, so a changeling counts.
// The Drover is not itself a Dinosaur.
func init() {
	whileDino := effects.SelfWhileYouControlA(effects.Subtype("Dinosaur"))
	effects.Register(effects.Spec{
		OracleID:     "6deba6e8-077b-4602-86af-64db766fbbd9",
		Name:         "Drover of the Mighty",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +2/+2 as long as you control a Dinosaur.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileDino,
			Apply:     effects.PumpBy(2, 2),
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
