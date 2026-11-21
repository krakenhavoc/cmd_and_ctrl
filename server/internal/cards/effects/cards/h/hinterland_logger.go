package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hinterland Logger // Timber Shredder — {1}{G} Creature — Human Werewolf
// 2/1 // Creature — Werewolf 4/2 (Innistrad Remastered, S82 GAP, #5948):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Trample
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers (innistrad_werewolves.go):
// each face prints its own intervening-"if" upkeep trigger on every
// player's upkeep. Trample is the back face's printed keyword.
//
// No simplification.
func init() {
	const oracle = "29b22d41-4c6e-46de-8553-983d05f12f0f"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Hinterland Logger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Hinterland Logger"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Timber Shredder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Timber Shredder"),
		},
	})
}
