package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Felisa, Fang of Silverquill — {2}{W}{B} Legendary Creature — Vampire Wizard
// 3/2
//
//	"Flying
//	 Mentor
//	 Whenever a nontoken creature you control dies, if it had counters on it,
//	 create X tapped 2/1 white and black Inkling creature tokens with flying,
//	 where X is the number of counters it had on it."
//
// Felisa's own death counts. The counters are read as the creature last
// existed on the battlefield (CR 603.10a, 608.2h) and every kind counts: the
// intervening "if" is that it had at least one, and X is the total.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "39144f72-789e-457a-a7e7-862f90c34e15",
		Name:            "Felisa, Fang of Silverquill",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Mentor("Felisa, Fang of Silverquill — mentor"),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !effects.ACreatureYouControlDied(ev, source, lki, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsToken() && effects.Cmdr2B25CounterTotal(g.LastKnownCountersForEffect(ev.CardID)) > 0
			}, "Felisa, Fang of Silverquill — create X tapped 2/1 Inkling tokens with flying, where X is the number of counters it had",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					info, ok := effects.NewContext(g, item).TriggeringPermanent()
					if !ok {
						return nil
					}
					x := 0
					for _, n := range info.Counters {
						x += n
					}
					if x <= 0 {
						return nil
					}
					tmpl := effects.TokenCard("2/1 white and black Inkling with flying")
					tmpl.Tapped = true
					return effects.CreateToken{Template: tmpl, N: x}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
