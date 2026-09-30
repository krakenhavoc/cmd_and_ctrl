package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meat Locker // Drowned Diner — Enchantment — Room (CR 709.5):
//
//	Meat Locker {2}{U}: "When you unlock this door, tap up to one target
//	creature and put two stun counters on it."
//	Drowned Diner {3}{U}{U}: "When you unlock this door, draw three cards,
//	then discard a card."
func init() {
	Register(Room(RoomSpec{
		OracleID:     "897eef47-3e99-4a5f-8a3e-ddb06bc95e9a",
		Name:         "Meat Locker // Drowned Diner",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{Targeting(
			WhenYouUnlockThisDoor(game.DoorLeft, "Meat Locker — tap up to one target creature and put two stun counters on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 2}.Apply(ctx)
					}
					return nil
				}),
			TargetCreature("up to one target creature").WithCount(0, 1),
		)}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Drowned Diner — draw three cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					if err := (DrawCards{Player: item.Controller, N: 3}).Apply(NewContext(g, item)); err != nil {
						return err
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: item.Controller, Source: item.SourceCardID, N: 1})
					return nil
				}),
		}},
	}))
}
