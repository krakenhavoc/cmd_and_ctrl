package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Villagers of Estwald // Howlpack of Estwald — {2}{G} Creature — Human
// Werewolf 2/3 // Creature — Werewolf 4/6 (Innistrad Remastered, S82 GAP,
// #5948):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers (innistrad_werewolves.go),
// the Hinterland Logger's without the trample.
//
// No simplification.
func init() {
	const oracle = "dcf692f6-baec-4b60-a6e3-2a8b731d7816"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Villagers of Estwald",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Villagers of Estwald"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Howlpack of Estwald",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Howlpack of Estwald"),
		},
	})
}
