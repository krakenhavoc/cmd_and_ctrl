package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earthlink — Enchantment {3}{B}{R}{G} (Ice Age, S82 OLD batch 846, #5636):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless you
//	 pay {2}.
//	 Whenever a creature dies, that creature's controller sacrifices a land of
//	 their choice."
//
// The upkeep line is Aura Flux's pay-or-sacrifice. The second is one trigger
// for each creature that dies, whoever controls it: the player is read from
// the creature as it last existed (CR 603.10a, 608.2h), and that player, not
// Earthlink's controller, chooses which of their lands to sacrifice
// (CR 701.21a); with no land they sacrifice nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6e4a5c6-3c9b-49d5-8288-1be46be657dc",
		Name:         "Earthlink",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Earthlink — sacrifice this enchantment unless you pay {2}", func(g *game.Game, item *game.StackItem) error {
				return effects.UpkeepPayUnless{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Earthlink — pay {2} or sacrifice this enchantment?",
					Else:     effects.UpkeepSacrificeSource,
				}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Earthlink — that creature's controller sacrifices a land of their choice",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					obj := ctx.Trigger().Object
					if obj == nil {
						return nil
					}
					info, ok := g.LastKnownPermanentForEffect(obj.Ref().ID)
					if !ok {
						return nil
					}
					lands := effects.LandsControlledByPlayer(g, info.Controller)
					if len(lands) == 0 {
						return nil
					}
					return effects.SacrificeChoice{Player: info.Controller, Candidates: lands, Question: "Earthlink — sacrifice a land"}.Apply(ctx)
				}),
		},
	})
}
