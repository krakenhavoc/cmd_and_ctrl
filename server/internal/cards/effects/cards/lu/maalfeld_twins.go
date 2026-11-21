package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maalfeld Twins — Creature — Zombie {5}{B}, 4/4 (Foundations, #4465):
//
//	"When this creature dies, create two 2/2 black Zombie creature
//	 tokens."
//
// Synapse Necromage's dies trigger with the plain Zombie token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dda4b515-49c0-43fe-9f6a-36defa326bb1",
		Name:         "Maalfeld Twins",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Maalfeld Twins — create two 2/2 black Zombie tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 2})),
		},
	})
}
