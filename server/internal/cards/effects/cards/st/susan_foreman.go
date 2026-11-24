package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Susan Foreman — Legendary Creature — Time Lord {1}{G}, 1/1 (Doctor Who):
//
//	"If you would planeswalk, instead look at the top two cards of your
//	 planar deck, put one on the bottom of your planar deck and the other
//	 on top, then planeswalk.
//	 {T}: Add {G}.
//	 Doctor's companion (You can have two commanders if the other is the
//	 Doctor.)"
//
// The first is a replacement of planeswalking (CR 614.1a, 701.31, #6096):
// game.ReplacePlaneswalkOrderTopTwo, read off the battlefield whenever her
// controller would planeswalk. In a Planechase game with one communal
// planar deck, "your planar deck" is that deck (CR 901.15c). Doctor's
// companion is a deck-building rule, read by the commander check
// (GapHasDoctorsCompanion).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f21f3b1-aee9-4272-ad02-067aa5f6a7c8",
		Name:         "Susan Foreman",
		Completeness: effects.CompletenessFull,
		GameActionReplacements: []game.GameActionReplacement{
			{Kind: game.ReplacePlaneswalkOrderTopTwo},
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "{T}: Add {G}.",
		}},
	})
}
