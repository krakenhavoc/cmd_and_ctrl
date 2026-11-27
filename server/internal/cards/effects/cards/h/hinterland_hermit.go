package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hinterland Hermit // Hinterland Scourge — Creature — Human Werewolf {1}{R}
// 2/1 // Creature — Werewolf 3/2:
//
//	Hinterland Hermit: "At the beginning of each upkeep, if no spells were
//	 cast last turn, transform this creature."
//	Hinterland Scourge: "This creature must be blocked if able.
//	 At the beginning of each upkeep, if a player cast two or more spells last
//	 turn, transform this creature."
func init() {
	const oracle = "8287306b-47d3-4e95-aa39-551f949d687c"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Hinterland Hermit",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.TransformAtUpkeepIfNoSpellsLastTurn("Hinterland Hermit")},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Hinterland Scourge",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.MustBeBlocked()},
		Triggered:    []game.TriggeredAbility{effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Hinterland Scourge")},
	})
}
