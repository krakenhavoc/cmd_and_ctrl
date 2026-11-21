package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Redcap Thief — Creature — Goblin Rogue {2}{R}, 2/3 (Wilds of Eldraine,
// #4471):
//
//	"When this creature enters, create a Treasure token."
//
// Contract Hero's enters trigger: one Treasure for its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e055bc6d-d395-4f85-8e15-3bce65b7fcca",
		Name:         "Redcap Thief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Redcap Thief — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
