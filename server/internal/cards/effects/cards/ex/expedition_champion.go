package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expedition Champion — Creature — Human Warrior {2}{R}, 2/3 (Foundations
// Jumpstart):
//
//	"This creature gets +2/+0 as long as you control another Warrior."
//
// A live layer 7c static on the source alone, gated on some other Warrior its
// controller controls, read through the layers (a changeling counts) and
// ending the moment the last one leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d432857f-cff6-421e-949b-ce2f65854e1d",
		Name:         "Expedition Champion",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +2/+0 as long as you control another Warrior.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.BnsB29SelfWhileAnother("Warrior"),
			Apply:     effects.PumpBy(2, 0),
		}},
	})
}
