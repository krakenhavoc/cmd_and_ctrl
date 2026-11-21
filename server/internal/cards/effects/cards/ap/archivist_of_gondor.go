package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archivist of Gondor — Creature — Human Advisor {2}{U}, 2/3 (Tales of
// Middle-earth Commander):
//
//	"When your commander deals combat damage to a player, if there is no
//	 monarch, you become the monarch.
//	 At the beginning of the monarch's end step, that player draws a card."
//
// "Your commander" is a card you own that is designated as a commander (CR
// 903.3), wherever it is controlled. The first clause has an intervening
// "if" (CR 603.4), checked as the damage is dealt and again on
// resolution. The second is an extra card for the monarch at the monarch's
// own end step, in addition to the monarch's draw the rules already give
// (CR 725.2); it fires on that step only, for whoever holds the crown as
// the step begins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a086f7a5-1319-4c5d-b23f-41bec2f81027",
		Name:         "Archivist of Gondor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(s82GapB118NoMonarch,
				effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
						return false
					}
					src, ok := g.LookupCardForEffect(ev.Source)
					return ok && src.IsCommander && src.Owner == source.Controller && g.Monarch == uuid.Nil
				}, "Archivist of Gondor — you become the monarch", effects.Do(effects.BecomeTheMonarch{}))),
			effects.On(game.EventStepBegan, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventStepBegan && ev.Step == game.StepEnd && ev.Actor != uuid.Nil && g.Monarch == ev.Actor
			}, "Archivist of Gondor — the monarch draws a card", func(g *game.Game, item *game.StackItem) error {
				who := effects.NewContext(g, item).Trigger().Event.Actor
				if who == uuid.Nil {
					return nil
				}
				return effects.DrawCards{Player: who, N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

var s82GapB118NoMonarch = effects.CondBoard{
	Name: "there is no monarch",
	Test: func(ctx *effects.Context) bool { return ctx.Game.Monarch == uuid.Nil },
}
