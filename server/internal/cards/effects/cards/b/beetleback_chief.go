package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beetleback Chief — Creature — Goblin Warrior {2}{R}{R}, 2/2 (EDHREC
// rank 3580):
//
//	"When this creature enters, create two 1/1 red Goblin creature
//	 tokens."
//
// Four power across three bodies. The tokens are Krenko's Goblins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5c5f64c-6911-430c-a825-b32b96d39c7d",
		Name:         "Beetleback Chief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Beetleback Chief — create two 1/1 red Goblins", effects.Do(effects.CreateToken{Template: effects.RedGoblinToken(), N: 2})),
		},
	})
}
