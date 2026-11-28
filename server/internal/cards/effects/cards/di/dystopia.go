package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dystopia — Enchantment {1}{B}{B} (Alliances, S82 OLD batch 818):
//
//	"Cumulative upkeep—Pay 1 life.
//	 At the beginning of each player's upkeep, that player sacrifices a green
//	 or white permanent of their choice."
//
// The upkeep's player is the trigger event's actor (CR 603.2), and the
// sacrifice is that player's own choice; a player with no green or white
// permanent sacrifices nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0e6203b-cef2-470f-8c79-7b139290e2d5",
		Name:         "Dystopia",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Dystopia — cumulative upkeep: pay 1 life", effects.LifePayment(1), "", ""),
			effects.AtEachUpkeep("Dystopia — that player sacrifices a green or white permanent of their choice",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					g.PlayerSacrificesForEffect(item.SourceCardID, ctx.Trigger().Event.Actor,
						effects.SacrificeSpec("a green or white permanent", effects.Or(effects.OfColor("G"), effects.OfColor("W"))),
						"Dystopia — sacrifice a green or white permanent")
					return nil
				}),
		},
	})
}
