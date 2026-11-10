package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vengeful Tracker — Creature — Human Detective {1}{R}, 2/2 (Murders at
// Karlov Manor, #4469):
//
//	"Whenever an opponent sacrifices an artifact, this creature deals 2
//	 damage to them."
//
// Nuka-Cola Vending Machine's watch on EventSacrifice, which fires before
// the zone move so the permanent is still an artifact to read (a token
// artifact and an artifact creature both count). The damage goes to the
// sacrificing player, named by the event's Actor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe540a23-0357-44fa-8e1c-0a5c2a04cb01",
		Name:         "Vengeful Tracker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventSacrifice || ev.Actor == source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsArtifact()
			}, "Vengeful Tracker — 2 damage to that opponent",
				Do(Instruction(func(ctx *Context) error {
					return DealDamage{Source: ctx.Item.SourceCardID, Target: ctx.Trigger().Event.Actor, Amount: 2}.Apply(ctx)
				}))),
		},
	})
}
