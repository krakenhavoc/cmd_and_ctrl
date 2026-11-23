package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meat Locker // Drowned Diner — Enchantment — Room (CR 709.5):
//
//	Meat Locker {2}{U}: "When you unlock this door, tap up to one target
//	creature and put two stun counters on it."
//	Drowned Diner {3}{U}{U}: "When you unlock this door, draw three cards,
//	then discard a card."
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "897eef47-3e99-4a5f-8a3e-ddb06bc95e9a",
		Name:         "Meat Locker // Drowned Diner",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenYouUnlockThisDoor(game.DoorLeft, "Meat Locker — tap up to one target creature and put two stun counters on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					targets := ctx.LegalTargets()
					if len(targets) == 0 {
						return nil
					}
					if err := (effects.TapTarget{Target: targets[0].ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.AddCounter{Target: targets[0].ID, Kind: game.CounterStun, N: 2}.Apply(ctx)
				}),
			effects.TargetCreature("up to one target creature").WithCount(0, 1),
		)}},
		Right: effects.Door{Triggered: []game.TriggeredAbility{
			effects.WhenYouUnlockThisDoor(game.DoorRight, "Drowned Diner — draw three cards, then discard a card",
				effects.Do(
					// The discard waits for a prompt the draws ask (CR 608.2c, #4155).
					effects.DrawCards{N: 3},
					effects.Instruction(func(ctx *effects.Context) error {
						ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: ctx.Item.Controller, Source: ctx.Item.SourceCardID, N: 1})
						return nil
					}),
				)),
		}},
	}))
}
