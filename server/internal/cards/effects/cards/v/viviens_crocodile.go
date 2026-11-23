package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vivien's Crocodile — Creature — Crocodile Spirit {2}{G}, 3/3 (Core Set 2020, #5518):
//
//	"This creature gets +1/+1 as long as you control a Vivien planeswalker."
//
// Briarbridge Tracker's conditional self pump (selfWhileYouControlA) over
// a planeswalker with the Vivien subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b49791f9-525c-4b04-954d-d7678c5565e3",
		Name:         "Vivien's Crocodile",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+1 as long as you control a Vivien planeswalker.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.Subtype("Vivien"))),
			Apply:     effects.PumpBy(1, 1),
		}},
	})
}
