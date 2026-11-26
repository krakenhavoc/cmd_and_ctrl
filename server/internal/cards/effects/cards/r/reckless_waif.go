package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Waif // Merciless Predator — {R} Creature — Human Rogue Werewolf 1/1 // Creature — Werewolf 3/2 (Innistrad,
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
	const oracle = "ba57df46-11c9-4f28-8818-175fb52e67bc"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Reckless Waif",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Reckless Waif"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Merciless Predator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Merciless Predator"),
		},
	})
}
