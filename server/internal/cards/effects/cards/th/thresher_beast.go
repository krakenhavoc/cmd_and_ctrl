package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thresher Beast — Creature — Beast {3}{G}{G}, 4/4 (Prophecy, S82 OLD b671, #5617):
//
//	"Whenever this creature becomes blocked, defending player sacrifices a land
//	 of their choice."
//
// Battle-Scarred Goblin's becomes-blocked trigger (once per blocked attacker,
// CR 509.3c), with Destructive Urge's land sacrifice: the defending player is the
// event's actor (the controller of the blockers) and chooses the land. A player
// with no land sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e88915fd-7867-45e6-b4b9-ca0af1823c0d",
		Name:         "Thresher Beast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Thresher Beast — defending player sacrifices a land of their choice",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
						return nil
					}
					g.PlayerSacrificesForEffect(item.SourceCardID, item.Trigger.Event.Actor, effects.SacrificeSpec("a land", effects.Land()),
						"Thresher Beast — sacrifice a land")
					return nil
				}),
		},
	})
}
