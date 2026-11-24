package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relic Robber — Creature — Goblin Rogue {2}{R}, 2/2 (Zendikar Rising,
// #5508):
//
//	"Haste
//	 Whenever this creature deals combat damage to a player, that player
//	 creates a 0/1 colorless Goblin Construct artifact creature token with
//	 'This token can't block' and 'At the beginning of your upkeep, this
//	 token deals 1 damage to you.'"
//
// The damaged player, not the Robber's controller, creates the token, so
// the upkeep ping hits them (Nettling Nuisance's shape). A player who has
// left the game gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "85eafa00-51df-4070-9b2d-247457ba3364",
		Name:            "Relic Robber",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Relic Robber — that player creates a 0/1 Goblin Construct that can't block and pings them each upkeep",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					victim := g.PlayerByIDForEffect(item.Trigger.Event.Target)
					if victim == nil || victim.Eliminated {
						return nil
					}
					return effects.CreateToken{Controller: victim.ID, Template: effects.S82OldB097GoblinConstructToken(), N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
