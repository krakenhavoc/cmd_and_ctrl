package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Solitary Hunter // One of the Pack — {3}{G} Creature — Human Warrior
// Werewolf 3/4 // Creature — Werewolf 5/6 (Shadows over Innistrad, S82 OLD
// batch 290, #5543):
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
	const oracle = "c99834dc-755f-407c-ad0e-550f8ae12198"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Solitary Hunter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Solitary Hunter"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "One of the Pack",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("One of the Pack"),
		},
	})
}
