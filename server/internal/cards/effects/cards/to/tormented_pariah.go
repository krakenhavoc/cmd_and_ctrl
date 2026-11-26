package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tormented Pariah // Rampaging Werewolf — {3}{R} Creature — Human Warrior Werewolf 3/2 // Creature — Werewolf 6/4 (Innistrad,
// S82 OLD b445, #5572):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last turn, transform this creature."
//	Back:  "At the beginning of each upkeep, if a player cast two or more spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers (innistrad_werewolves.go):
// each face prints its own intervening-"if" upkeep trigger on every player's
// upkeep.
//
// No simplification.
func init() {
	const oracle = "bbcc2d2d-b29c-410f-bee7-5ef1609a5b32"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Tormented Pariah",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Tormented Pariah"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Rampaging Werewolf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Rampaging Werewolf"),
		},
	})
}
