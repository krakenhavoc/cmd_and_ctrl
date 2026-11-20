package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mayor of Avabruck // Howlpack Alpha — {1}{G} Creature — Human Advisor
// Werewolf 1/1 // Creature — Werewolf 3/3 (#4071, ADR 0132 amendment
// 2026-10-16):
//
//	Front: "Other Human creatures you control get +1/+1.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Each other creature you control that's a Werewolf or a Wolf
//	        gets +1/+1.
//	        At the beginning of your end step, create a 2/2 green Wolf
//	        creature token.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// Both lords are the shared tribal anthem over your OTHER creatures of
// the type, so turning over swaps which tribe gets the bonus in the same
// layer pass. The end-step token is the Alpha's controller's only.
//
// No simplification.
func init() {
	const oracle = "6e638587-30b3-4c4b-b463-fb415ea048f7"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Mayor of Avabruck",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Human"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Mayor of Avabruck"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Howlpack Alpha",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Werewolf", "Wolf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Howlpack Alpha — create a 2/2 green Wolf creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1})),
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Howlpack Alpha"),
		},
	})
}
