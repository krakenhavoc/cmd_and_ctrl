package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mage Hunter — {3}{B} Creature — Horror 3/4 (Strixhaven: School of Mages,
// #5502):
//
//	"Whenever an opponent casts or copies an instant or sorcery spell, they
//	 lose 1 life."
//
// Magecraft's two events (a cast, and a copy put on the stack, CR 707.10)
// read from the opponent's side: the spell is an instant or sorcery and the
// player who cast or copied it is not the Mage Hunter's controller. "They"
// is that player, read off the item's carried event, so a copy someone else
// makes costs them life, not the Mage Hunter's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3d2df4d-8c79-4391-bc4e-69325958a5be",
		Name:         "Mage Hunter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny(effects.MagecraftEvents, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (spell.IsInstant() || spell.IsSorcery())
			}, "Mage Hunter — that player loses 1 life", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				caster := ctx.Trigger().Event.Actor
				if caster == uuid.Nil {
					return nil
				}
				return effects.GainLife{Player: caster, Amount: -1}.Apply(ctx)
			}),
		},
	})
}
