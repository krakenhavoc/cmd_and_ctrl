package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Marshal — Creature — Goblin Warrior, {4}{R}{R}, 3/3:
//
//	"Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters or dies, create two 1/1 red Goblin creature tokens."
//
// "Enters or dies" is one ability with two trigger conditions.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4bc91d7d-37aa-475c-8fe9-763975d51add",
		Name:         "Goblin Marshal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Goblin Marshal", "{4}{R}{R}"),
			effects.WhenThisEntersOrDies("Goblin Marshal — create two 1/1 red Goblins", effects.Do(effects.CreateToken{Template: effects.RedGoblinToken(), N: 2})),
		},
	})
}
