package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maalfeld Twins — Creature — Zombie {5}{B}, 4/4 (Foundations, #4465):
//
//	"When this creature dies, create two 2/2 black Zombie creature
//	 tokens."
//
// Synapse Necromage's dies trigger with the plain Zombie token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dda4b515-49c0-43fe-9f6a-36defa326bb1",
		Name:         "Maalfeld Twins",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Maalfeld Twins — create two 2/2 black Zombie tokens",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 2})),
		},
	})
}
