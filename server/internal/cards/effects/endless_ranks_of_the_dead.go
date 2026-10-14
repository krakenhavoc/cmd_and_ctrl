package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Endless Ranks of the Dead — Enchantment {2}{B}{B}:
//
//	"At the beginning of your upkeep, create X 2/2 black Zombie
//	 creature tokens, where X is half the number of Zombies you
//	 control, rounded down."
//
// X is counted as the trigger resolves (CR 608.2h), over every Zombie
// PERMANENT you control, not only creatures: the oracle says
// "Zombies". One Zombie makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69d4ecac-4735-4667-bfc1-c8800b436d08",
		Name:         "Endless Ranks of the Dead",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Endless Ranks of the Dead — create a Zombie for every two Zombies you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := zombiesYouControl(g, item.Controller) / 2
					return CreateToken{Template: TokenCard("2/2 black Zombie"), N: n}.Apply(ctx)
				}),
		},
	})
}

// zombiesYouControl is "the number of Zombies you control": every
// permanent `player` controls with the Zombie subtype.
func zombiesYouControl(g *game.Game, player uuid.UUID) int {
	return len(permanentsControlledByMatching(g, player, HasSubtype("Zombie")))
}
