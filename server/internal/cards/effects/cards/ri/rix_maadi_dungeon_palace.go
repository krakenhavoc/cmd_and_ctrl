package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rix Maadi, Dungeon Palace — Land (Dissension, #5598):
//
//	"{T}: Add {C}.
//	 {1}{B}{R}, {T}: Each player discards a card. Activate only as a sorcery."
//
// The discard is Mindlash Sliver's tspB1EachPlayerDiscardsACard: every
// player, in APNAP order, picks a card of their own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1bee68d-135b-4e30-8830-48a3315d13a9",
		Name:         "Rix Maadi, Dungeon Palace",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}{R}, {T}: Each player discards a card. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{B}{R}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect:       effects.TspB1EachPlayerDiscardsACard("Rix Maadi, Dungeon Palace — discard a card"),
		}},
	})
}
