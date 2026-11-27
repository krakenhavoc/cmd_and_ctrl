package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sacred Prey — Creature — Horse {G}, 1/1 (Mercadian Masques, #5619):
//
//	"Whenever this creature becomes blocked, you gain 1 life."
//
// Battle-Scarred Goblin's trigger shape: "becomes blocked" is once per
// blocked attacker however many creatures block it (CR 509.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e24110e8-72cf-4dd0-bca0-665758daca88",
		Name:         "Sacred Prey",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Sacred Prey — you gain 1 life",
				effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
