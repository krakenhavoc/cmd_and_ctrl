package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Colleen Wing, Street Samurai — Legendary Creature — Human Samurai Hero
// {1}{W}, 2/2 (Marvel Super Heroes, #4455):
//
//	"Whenever you cast a spell that targets a creature you control, put a
//	 +1/+1 counter on Colleen Wing. Scry 1."
//
// Heroic's shape (shield_families_recipient.go) widened from "this creature"
// to any creature the controller controls: the spell's targets are chosen
// before it becomes cast (CR 601.2c), so they are on its stack item as the
// cast event is emitted. A spell with several such targets triggers it
// once; an ability is not a spell and a copy is not cast, so neither
// triggers it. The counter goes on Colleen (unless she is a new object by
// resolution, CR 400.7), then the scry happens either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "685181dd-0124-4c76-9270-f9dd17e6a221",
		Name:         "Colleen Wing, Street Samurai",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				item := g.StackItemForEffect(ev.CardID)
				if item == nil || item.Kind != game.StackItemSpell {
					return false
				}
				for _, t := range item.Targets {
					if t.Kind != game.TargetCard {
						continue
					}
					if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() && c.Controller == source.Controller && onBattlefield(g, t.ID) {
						return true
					}
				}
				return false
			}, "Colleen Wing, Street Samurai — a +1/+1 counter on her, then scry 1", Do(
				Instruction(func(ctx *Context) error {
					if sourceIsNewObject(ctx.Game, ctx.Item) || !onBattlefield(ctx.Game, ctx.Source()) {
						return nil
					}
					return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
				Scry{N: 1},
			)),
		},
	})
}
