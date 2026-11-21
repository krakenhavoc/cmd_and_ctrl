package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Gang Leader — Creature — Goblin Warrior {2}{R}{R}, 2/2 (Mystery
// Booster 2):
//
//	"When this creature enters, create two 1/1 red Goblin creature tokens."
//
// One enters-the-battlefield trigger, two tokens made together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79d99305-b3dd-4687-af69-bf4b28704aae",
		Name:         "Goblin Gang Leader",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Goblin Gang Leader — create two 1/1 red Goblin creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin"), N: 2})),
		},
	})
}
