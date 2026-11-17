package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathcurse Ogre — Creature — Ogre Warrior {5}{B}, 3/3 (Champions of
// Kamigawa, #5603):
//
//	"When this creature dies, each player loses 3 life."
//
// A dies trigger (CR 700.4). Every player still in the game loses 3,
// its controller first and then the opponents in seat order; a loss, not
// damage, so no prevention shield sees it (CR 119.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d40bc98-28e7-4fa5-b848-1cbac5345c12",
		Name:         "Deathcurse Ogre",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Deathcurse Ogre — each player loses 3 life", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range tablePlayers(ctx) {
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), id, -3); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
