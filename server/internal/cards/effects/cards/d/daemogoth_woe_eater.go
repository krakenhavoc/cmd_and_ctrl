package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daemogoth Woe-Eater — Creature — Demon {1}{B}{B/G}{G}, 7/6 (Strixhaven:
// School of Mages, #5502):
//
//	"At the beginning of your upkeep, sacrifice a creature.
//	 When you sacrifice this creature, each opponent discards a card, you draw
//	 a card, and you gain 2 life."
//
// The upkeep sacrifice is mandatory and may be the Woe-Eater itself (Shadowborn
// Demon's PlayerSacrificesForEffect). The second ability is a trigger of its
// own on the sacrifice event (Biolume Egg's): it fires when its controller
// sacrifices the Woe-Eater by any means, the upkeep trigger or another
// permanent's cost or effect, and not when it dies any other way. Each
// opponent chooses their own discard (EachOpponentDiscardsOne); the draw and
// the life follow, in that order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "290323d7-d8cb-4706-87d9-b920a4e2f680",
		Name:         "Daemogoth Woe-Eater",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Daemogoth Woe-Eater — sacrifice a creature",
				func(g *game.Game, item *game.StackItem) error {
					g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
						effects.SacrificeSpec("a creature", effects.Creature()), "Daemogoth Woe-Eater — sacrifice a creature")
					return nil
				}),
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor == source.Controller
			}, "Daemogoth Woe-Eater — each opponent discards a card, you draw a card, and you gain 2 life",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
					effects.DrawCards{N: 1},
					effects.GainLife{Amount: 2},
				)),
		},
	})
}
