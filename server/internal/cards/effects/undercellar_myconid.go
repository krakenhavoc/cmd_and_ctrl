package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undercellar Myconid — Creature — Fungus {2}{G}, 1/2 (Commander Legends:
// Battle for Baldur's Gate, #5490):
//
//	"Whenever this creature enters or dies, create a 1/1 green Saproling
//	 creature token.
//	 {T}: Add one mana of any color."
//
// Mogg War Marshal's enters-or-dies row with the table's Saproling, and
// Birds of Paradise's mana ability. The mana ability needs the creature
// to have been under your control since your most recent turn began
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "127f5b7e-79bc-4fe0-b9d4-1a267a6a4d46",
		Name:         "Undercellar Myconid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Undercellar Myconid — create a 1/1 green Saproling creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
