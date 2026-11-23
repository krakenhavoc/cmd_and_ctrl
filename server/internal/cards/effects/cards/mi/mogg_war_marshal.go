package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg War Marshal — Creature — Goblin Warrior, {1}{R}, 1/1:
//
//	"Echo {1}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters or dies, create a 1/1 red Goblin creature token."
//
// "Enters or dies" is one ability with two trigger conditions.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b8f4879-578e-40e4-863a-41d0c32c6bdd",
		Name:         "Mogg War Marshal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Mogg War Marshal", "{1}{R}"),
			effects.WhenThisEntersOrDies("Mogg War Marshal — create a 1/1 red Goblin", effects.Do(effects.CreateToken{Template: effects.RedGoblinToken(), N: 1})),
		},
	})
}
