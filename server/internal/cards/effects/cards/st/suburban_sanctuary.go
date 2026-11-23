package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "6ef30340-a26d-49aa-bc86-0b8aa5252f87",
		Name:          "Suburban Sanctuary",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("G", "W")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
