package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Village Ironsmith // Ironfang — {1}{R} Creature — Human Werewolf 1/1 // Creature — Werewolf 3/1 (Innistrad,
// S82 OLD b445, #5572):
//
//	Front: "First strike
//	        At the beginning of each upkeep, if no spells were cast last turn, transform this creature."
//	Back:  "First strike
//	        At the beginning of each upkeep, if a player cast two or more spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers (innistrad_werewolves.go):
// each face prints its own intervening-"if" upkeep trigger on every player's
// upkeep.
//
// No simplification.
func init() {
	const oracle = "c7b1aaff-e2c3-427c-8478-1fd053f7ffb2"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Village Ironsmith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Village Ironsmith"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Ironfang",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Ironfang"),
		},
	})
}
