package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sengir Autocrat — {3}{B} Creature — Human 2/2 (Eternal Masters, S82 OLD,
// #5542):
//
//	"When this creature enters, create three 0/1 black Serf creature tokens.
//	 When this creature leaves the battlefield, exile all Serf tokens."
//
// Two printed triggers. The leaves trigger exiles every Serf token on the
// battlefield, whoever controls it (not just the ones this made), when the
// Autocrat leaves by any route.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0561043a-3896-4bb1-8ff5-ee674ddfea2a",
		Name:         "Sengir Autocrat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sengir Autocrat — create three 0/1 black Serf creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 black Serf"), N: 3})),
			effects.WhenThisLeaves("Sengir Autocrat — exile all Serf tokens",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsToken() && c.HasSubtypeWord("Serf")
					}}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
