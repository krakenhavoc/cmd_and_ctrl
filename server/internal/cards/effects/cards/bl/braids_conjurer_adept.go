package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Braids, Conjurer Adept — Legendary Creature — Human Wizard {2}{U}{U}, 2/2
// (Commander Legends: Battle for Baldur's Gate):
//
//	"At the beginning of each player's upkeep, that player may put an
//	 artifact, creature, or land card from their hand onto the battlefield."
//
// Gate to the Aether's each-upkeep trigger (the event's actor is the
// upkeep's player), with the put-from-hand pick made by, and the permanent
// entering under the control of, that player rather than Braids' controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "359b8f2c-831e-47d0-9267-dae79cecb63a",
		Name:         "Braids, Conjurer Adept",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Braids, Conjurer Adept — that player may put an artifact, creature, or land card from their hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					player := item.Trigger.Event.Actor
					if g.PlayerByIDForEffect(player) == nil {
						return nil
					}
					return effects.PutFromHandOntoBattlefield{
						Player:   player,
						Match:    effects.Or(effects.Artifact(), effects.Creature(), effects.Land()),
						Optional: true,
						Label:    "Braids, Conjurer Adept — you may put an artifact, creature, or land card from your hand onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
