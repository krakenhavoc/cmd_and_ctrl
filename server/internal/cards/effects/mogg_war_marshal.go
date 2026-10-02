package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mogg War Marshal — Creature — Goblin Warrior, {1}{R}, 1/1:
//
//	"Echo {1}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters or dies, create a 1/1 red Goblin creature token."
//
// "Enters or dies" is one ability with two trigger conditions.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b8f4879-578e-40e4-863a-41d0c32c6bdd",
		Name:         "Mogg War Marshal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Mogg War Marshal", "{1}{R}"),
			WhenThisEntersOrDies("Mogg War Marshal — create a 1/1 red Goblin", Do(CreateToken{Template: RedGoblinToken(), N: 1})),
		},
	})
}
