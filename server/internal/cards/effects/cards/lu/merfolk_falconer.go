package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merfolk Falconer — Creature — Merfolk Wizard {3}{U}{U}, 4/4 (Zendikar
// Rising, #5508):
//
//	"Flying
//	 Whenever you cast a kicked spell, scry 2."
//
// A cast trigger gated on the spell's own record of what was paid (CR 702.33,
// game.KickedTimesPaid over the stack item): kicker and multikicker both
// count, a spell cast without paying a kicker does not, and the trigger is the
// Falconer's controller's, for their own casts only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "77d5d435-d794-46c0-b630-a460973cf75a",
		Name:            "Merfolk Falconer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && game.KickedTimesPaid(c, g.StackItemPaidForEffect(ev.CardID).OptionalCosts) > 0
			}, "Merfolk Falconer — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
