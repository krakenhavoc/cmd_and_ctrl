package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reflections of Littjara — Enchantment {4}{U}:
//
//	"As this enchantment enters, choose a creature type.
//	 Whenever you cast a spell of the chosen type, copy that spell. (A
//	 copy of a permanent spell becomes a token.)"
//
// The type is a CR 614.12 as-enters choice stamped on the permanent
// (Card.NamedTribe, Banner of Kinship's shape). The cast trigger is
// Echoes of Eternity's: a CopySpell of the spell that triggered it, with
// the new-targets choice offered (CR 707.10c). "A spell of the chosen
// type" is the creature type on the spell as cast, so a changeling and
// a kindred spell of the type count. The copy is not cast, so it does
// not trigger the Reflections again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3fdfb94-2d10-4743-864c-a59fdd57d8b7",
		Name:         "Reflections of Littjara",
		Completeness: CompletenessFull,
		AsEnters: func(card *game.Card, ctx *Context) error {
			self := card.InstanceID
			ChooseCreatureTypeThen(ctx.Game, card.Controller, self, "Reflections of Littjara — choose a creature type",
				func(g *game.Game, t string) error {
					if t != "" {
						g.SetNamedTribeForEffect(self, t)
					}
					return nil
				})
			return nil
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || source.NamedTribe == "" {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.HasSubtype(source.NamedTribe)
			}, "Reflections of Littjara — copy that spell", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CopySpell{StackID: ctx.Trigger().Event.CardID, ChooseNewTargets: true}.Apply(ctx)
			}),
		},
	})
}
