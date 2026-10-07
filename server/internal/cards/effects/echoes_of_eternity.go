package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Echoes of Eternity — Kindred Enchantment — Eldrazi {3}{C}{C}{C}:
//
//	"If a triggered ability of a colorless spell you control or another
//	 colorless permanent you control triggers, that ability triggers an
//	 additional time.
//	 Whenever you cast a colorless spell, copy it. You may choose new
//	 targets for the copy. (A copy of a permanent spell becomes a
//	 token.)"
//
// The first line is a trigger doubler (Panharmonicon's
// shape) over the SOURCE of the ability: a colorless spell you control
// (the source is on the stack) or a colorless permanent you control
// other than Echoes itself — Echoes is colorless too, but "another"
// excludes it. Colour is read on the source's layered characteristics.
// The second line is a cast trigger with a CopySpell body, Jin-
// Gitaxias's shape, and the new-targets choice is offered (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23a79523-4be0-4d17-80aa-5ea024cb3463",
		Name:         "Echoes of Eternity",
		Completeness: CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{{
			Label: "Echoes of Eternity",
			Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
				if q.Ability == nil || q.SourceLKI.Controller != q.DoublerLKI.Controller {
					return false
				}
				if !q.FromSpell && q.Source.InstanceID == q.Doubler.InstanceID {
					return false
				}
				return len(q.SourceLKI.Colors) == 0
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsColorless()
			}, "Echoes of Eternity — copy that spell", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CopySpell{StackID: ctx.Trigger().Event.CardID, ChooseNewTargets: true}.Apply(ctx)
			}),
		},
	})
}
