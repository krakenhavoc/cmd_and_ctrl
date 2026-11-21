package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eloise, Nephalia Sleuth — Legendary Creature — Human Rogue {3}{U}{B},
// 4/4 (Midnight Hunt Commander, #5497):
//
//	"Whenever another creature you control dies, investigate.
//	 Whenever you sacrifice a token, surveil 1."
//
// Two triggers, one each. The first is a creature of yours dying other
// than Eloise (Blood Artist's reading of "dies" with the controller read
// from last-known information, CR 603.10a) and makes a Clue. The second
// is Ashling's reading of a sacrifice: the event is announced while the
// permanent is still on the battlefield, so "token" is read off it
// there. Sacrificing a Clue token to draw therefore surveils, and the
// sacrifice of a nontoken does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65e65eab-bc3f-4eae-8361-7f0a65fc1673",
		Name:         "Eloise, Nephalia Sleuth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ev.CardID != source.InstanceID && effects.ACreatureYouControlDied(ev, source, lki, g)
			}, "Eloise, Nephalia Sleuth — investigate", effects.Do(effects.Investigate{N: 1})),
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || ev.CardID == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsToken()
			}, "Eloise, Nephalia Sleuth — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
