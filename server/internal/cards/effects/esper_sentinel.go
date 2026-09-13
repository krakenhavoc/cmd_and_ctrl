package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Esper Sentinel — 1/1 Artifact Creature — Human Soldier for {W}:
//
//	"Whenever an opponent casts their first noncreature spell each
//	turn, draw a card unless that player pays {X}, where X is Esper
//	Sentinel's power."
//
// S19 sub-PR 6. Two engine pieces make this card:
//   - Game.SpellsCastThisTurn, bumped in CastSpell BEFORE EventCast
//     fires, so AppliesTo sees Noncreature == 1 for exactly the
//     first noncreature spell an opponent casts each turn.
//   - The pay-unless prompt, with X read at resolution from the
//     Sentinel's current power (post-layer, counters included — the
//     same number combat uses). If the Sentinel has left the
//     battlefield by then, X is the power captured when the trigger
//     went on the stack (CR 603.10). X of 0 is a free "payment": no
//     prompt, no draw.
func init() {
	Register(Spec{
		OracleID:     "5def9f38-0a0b-4e8d-9f9d-29dcb46520b4",
		Name:         "Esper Sentinel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || spell.IsCreature() {
					return false
				}
				return g.CastTallyFor(ev.Actor).Noncreature == 1
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				caster := ev.Actor
				powerAtTrigger := source.CurrentPower()
				return game.NewTriggeredItem(source, "Esper Sentinel — draw unless caster pays {X}",
					func(g *game.Game, item *game.StackItem) error {
						x := powerAtTrigger
						if live, ok := g.LookupCardForEffect(item.SourceCardID); ok {
							x = live.CurrentPower()
						}
						if x <= 0 {
							return nil
						}
						cost := fmt.Sprintf("{%d}", x)
						return PayUnless{
							Chooser:  caster,
							Cost:     cost,
							Question: "Esper Sentinel — pay " + cost + "?",
							OnDecline: func(ctx *Context) error {
								return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
							},
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
