package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gatstaf Arsonists // Gatstaf Ravagers — {4}{R} Creature — Human Werewolf
// 5/4 // Creature — Werewolf 6/5 (Shadows over Innistrad, S82 OLD batch 290,
// #5543):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Menace
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers (innistrad_werewolves.go);
// menace is the back face's printed keyword.
//
// No simplification.
func init() {
	const oracle = "cde622fa-a634-409d-962c-1e4c5ddf9b3b"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Gatstaf Arsonists",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Gatstaf Arsonists"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Gatstaf Ravagers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Gatstaf Ravagers"),
		},
	})
}
