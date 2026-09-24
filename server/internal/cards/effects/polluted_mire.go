package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polluted Mire — Land (EDHREC rank 4220):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The black member of Urza's Saga's cycling-land cycle. Nobody plays
// one for the mana — a tapped Swamp is a downgrade — they play it
// because a land that is a spell when you are flooded is worth a
// slot, and because it feeds a Life from the Loam or a Ramunap
// Excavator.
//
// Cycling {2} arrived with #660, the same activated ability Ketria
// Triome and Raffine's Tower carry — Cycling("{2}").
func init() {
	Register(Spec{
		OracleID:     "9809d975-7ef8-4946-9041-607c4e954b13",
		Name:         "Polluted Mire",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
