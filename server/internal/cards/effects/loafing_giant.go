package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "038dce4c-f754-45ef-98a4-4e16f931a65c",
		Name:         "Loafing Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
			}, "Loafing Giant — mill a card; if it's a land, prevent its combat damage this turn",
				func(g *game.Game, item *game.StackItem) error {
					return MillToZone{
						Player: item.Controller,
						N:      1,
						Then: func(ctx *Context, milled []uuid.UUID) error {
							for _, id := range milled {
								if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
									return shieldAgainstThisCombatDamage(ctx)
								}
							}
							return nil
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
