package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Loafing Giant — Creature — Giant {4}{R}, 4/6:
//
//	"Whenever this creature attacks or blocks, mill a card. If a land card was milled this way, prevent all combat damage this creature would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): one ability with two trigger
// conditions (Savvy Hunter's shape); a block triggers once however many
// attackers it blocks (CR 509.3a, its ruling). "Milled this way" is the
// card that reached the graveyard (MillToZone's Then), and if it is a
// land the Giant's own combat damage is prevented for the rest of the
// turn, the Giant being the shield's source while it is the same object
// (CR 400.7).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "038dce4c-f754-45ef-98a4-4e16f931a65c",
		Name:         "Loafing Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, "Loafing Giant — mill a card; if it's a land, prevent its combat damage this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MillToZone{
						Player:   item.Controller,
						N:        1,
						ThenTail: effects.NewContext(g, item).ItemTail(loafingGiantMilled),
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// loafingGiantMilled is the clause after the mill (#4253 slice 9).
var loafingGiantMilled = effects.CardTail("loafing-giant/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	milled := o.IDs

	for _, id := range milled {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
			return effects.ShieldAgainstThisCombatDamage(ctx)
		}
	}
	return nil
})
