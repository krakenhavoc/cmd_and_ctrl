package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grizzled Outcasts // Krallenhorde Wantons — {4}{G} Creature — Human Werewolf 4/4 // Creature — Werewolf 7/7 (Innistrad,
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
	const oracle = "abcfcc71-77c0-46e7-93c3-726967946b12"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Grizzled Outcasts",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Grizzled Outcasts"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Krallenhorde Wantons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Krallenhorde Wantons"),
		},
	})
}
