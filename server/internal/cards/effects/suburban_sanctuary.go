package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suburban Sanctuary — Land (Marvel's Spider-Man, #4460):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}.
//	 {4}, {T}: Surveil 1. (Look at the top card of your library. You may put
//	 it into your graveyard.)"
//
// Hedge Maze's tapped two-colour land with the surveil moved from the enters
// trigger to a {4}, {T} activation. It has no basic land type, so its mana
// ability is declared. The surveil is an activated ability, so it uses the
// stack and the land stays tapped to pay for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "6ef30340-a26d-49aa-bc86-0b8aa5252f87",
		Name:          "Suburban Sanctuary",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "W")},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
