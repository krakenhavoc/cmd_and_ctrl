package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Convicted Killer // Branded Howler — {2}{R} Creature — Human Werewolf 2/2
// // Creature — Werewolf 4/4 (Shadows over Innistrad, S82 OLD batch 290,
// #5543):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers (innistrad_werewolves.go).
//
// No simplification.
func init() {
	const oracle = "c1801b5c-1a87-46d0-83d3-e8dfd94211f4"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Convicted Killer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Convicted Killer"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Branded Howler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Branded Howler"),
		},
	})
}
