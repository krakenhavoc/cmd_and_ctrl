package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Redcap Thief — Creature — Goblin Rogue {2}{R}, 2/3 (Wilds of Eldraine,
// #4471):
//
//	"When this creature enters, create a Treasure token."
//
// Contract Hero's enters trigger: one Treasure for its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e055bc6d-d395-4f85-8e15-3bce65b7fcca",
		Name:         "Redcap Thief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Redcap Thief — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
