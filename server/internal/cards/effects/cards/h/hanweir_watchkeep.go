package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hanweir Watchkeep // Bane of Hanweir — {2}{R} Creature — Human Warrior
// Werewolf 1/5 // Creature — Werewolf 5/5 (Innistrad Remastered, S82 GAP,
// #5948):
//
//	Front: "Defender
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "This creature attacks each combat if able.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers
// (innistrad_werewolves.go). Defender is a printed keyword of the front,
// and the back's forced attack is the shared AttacksEachCombat static
// (CR 508.1d).
//
// No simplification.
func init() {
	const oracle = "8bdf986e-a47e-47b6-8cdb-576c86c4041b"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Hanweir Watchkeep",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Hanweir Watchkeep"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Bane of Hanweir",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Bane of Hanweir"),
		},
	})
}
