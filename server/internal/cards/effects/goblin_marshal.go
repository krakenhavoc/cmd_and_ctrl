package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Marshal — Creature — Goblin Warrior, {4}{R}{R}, 3/3:
//
//	"Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters or dies, create two 1/1 red Goblin creature tokens."
//
// "Enters or dies" is one ability with two trigger conditions.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bc91d7d-37aa-475c-8fe9-763975d51add",
		Name:         "Goblin Marshal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Goblin Marshal", "{4}{R}{R}"),
			WhenThisEntersOrDies("Goblin Marshal — create two 1/1 red Goblins", Do(CreateToken{Template: RedGoblinToken(), N: 2})),
		},
	})
}
