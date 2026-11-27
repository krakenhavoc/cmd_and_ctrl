package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Firebombers — Creature — Human Soldier {3}{R}{R}, 3/3 (Prophecy, S82
// OLD batch 674, #5617):
//
//	"When this creature enters, each player sacrifices all lands they control
//	 except for three."
//
// Magus of the Balance's sacrifice walk (PlayersSacrificeForEffect): each
// player with more than three lands is asked to sacrifice the excess, of their
// own choice (choosing the three to keep is choosing the rest to lose), the
// players answering in APNAP order (CR 101.4), and everything chosen goes
// together. A player with three or fewer lands is not asked and sacrifices
// nothing. The count is read as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5e02fa5-827e-414b-acb5-932de4bb3ee3",
		Name:         "Keldon Firebombers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Keldon Firebombers — each player sacrifices all lands they control except for three",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var asks []uuid.UUID
					for _, id := range effects.ApnapPlayers(g) {
						if p := ctx.PlayerByID(id); p == nil || p.Eliminated {
							continue
						}
						for i := effects.CountControlled(g, id, game.Card.IsLand); i > 3; i-- {
							asks = append(asks, id)
						}
					}
					if len(asks) == 0 {
						return nil
					}
					g.PlayersSacrificeForEffect(ctx.Source(), asks, effects.SacrificeSpec("a land", effects.Land()),
						"Keldon Firebombers — sacrifice all lands except for three")
					return nil
				}),
		},
	})
}
