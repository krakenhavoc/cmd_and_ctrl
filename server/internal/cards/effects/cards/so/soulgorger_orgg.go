package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulgorger Orgg — Creature — Nightmare Orgg {3}{R}{R}, 6/6 (Judgment):
//
//	"Trample
//	 When this creature enters, you lose all but 1 life.
//	 When this creature leaves the battlefield, you gain life equal to the
//	 life you lost when it entered."
//
// A life loss that names the Orgg as its source (CR 119.3), for all but one of
// the controller's life as it resolves (nothing at 1 or less). The leaves
// trigger reads the loss back off the event log: the life that changed because
// of THIS object's enters trigger, up to the object's own entry (CR 400.7), so
// a replaced or partial loss gives back exactly what was lost, and an Orgg
// that left before its enters trigger resolved gains nothing.
//
// No simplification.
func init() {
	const name = "Soulgorger Orgg"
	effects.Register(effects.Spec{
		OracleID:        "cbbcd3a3-6ddd-4511-a352-3055d0102e7c",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — you lose all but 1 life", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				p := g.PlayerByIDForEffect(ctx.Controller())
				if p == nil || p.Life <= 1 {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, ctx.Controller(), -(p.Life - 1))
			}),
			effects.WhenThisLeaves(name+" — you gain life equal to the life you lost when it entered", func(g *game.Game, item *game.StackItem) error {
				lost := S82OldB628LifeLostByEntering(g, item.SourceCardID, item.Controller)
				if lost <= 0 {
					return nil
				}
				return effects.GainLife{Player: item.Controller, Amount: lost}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// S82OldB628LifeLostByEntering is the life `player` lost because of the
// object `source` since it last entered the battlefield: the sum of the life
// changes in the event log that name `source` as their source, walked back
// from the newest to the object's latest entry onto the battlefield (CR
// 400.7: an earlier entry is another object).
func S82OldB628LifeLostByEntering(g *game.Game, source, player uuid.UUID) int {
	lost := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventZoneMove && ev.CardID == source && ev.NewZone == game.ZoneBattlefield {
			break
		}
		if ev.Kind == game.EventChangeLife && ev.Source == source && ev.Target == player && ev.Amount < 0 {
			lost += -ev.Amount
		}
	}
	return lost
}
