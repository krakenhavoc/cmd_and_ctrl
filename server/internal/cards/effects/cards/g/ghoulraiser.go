package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghoulraiser — Creature — Zombie {1}{B}{B}, 2/2 (Jumpstart 2022, #5948):
//
//	"When this creature enters, return a Zombie card at random from your
//	 graveyard to your hand."
//
// Haunted Fengraf's random pick on the keyed stream (an undo replays it)
// over the Zombie cards of the controller's graveyard; with none, nothing
// returns. "Zombie card" reads the card's subtype words, so a changeling
// counts (CR 205.3m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "614f748b-5d31-440d-a7c0-3ba12dc63c24",
		Name:         "Ghoulraiser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ghoulraiser — return a Zombie card at random from your graveyard to your hand", func(g *game.Game, item *game.StackItem) error {
				ids := effects.GraveyardIDs(g, item.Controller, func(c game.Card) bool { return c.HasSubtypeWord("Zombie") })
				pick := g.ChooseAtRandomForEffect(effects.RandomDraw(effects.NewContext(g, item)), ids, 1)
				if len(pick) == 0 {
					return nil
				}
				return g.ReturnFromGraveyardForEffect(pick[0], game.ZoneHand)
			}),
		},
	})
}
