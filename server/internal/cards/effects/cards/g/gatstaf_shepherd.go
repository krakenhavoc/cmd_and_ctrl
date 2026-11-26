package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gatstaf Shepherd // Gatstaf Howler — {1}{G} Creature — Human Werewolf 2/2 // Creature — Werewolf 3/3 (Innistrad,
// S82 OLD b445, #5572):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last turn, transform this creature."
//	Back:  "Intimidate
//	        At the beginning of each upkeep, if a player cast two or more spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers (innistrad_werewolves.go):
// each face prints its own intervening-"if" upkeep trigger on every player's
// upkeep.
//
// No simplification.
func init() {
	const oracle = "c05c5375-d0d7-4cf2-92e6-76f86cd0febf"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Gatstaf Shepherd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Gatstaf Shepherd"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Gatstaf Howler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Gatstaf Howler"),
		},
	})
}
