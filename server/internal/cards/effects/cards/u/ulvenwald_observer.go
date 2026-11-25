package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ulvenwald Observer — Creature — Treefolk {4}{G}{G}, 6/6 (Eldritch Moon,
// S82 OLD, #5541):
//
//	"Whenever a creature you control with toughness 4 or greater dies, draw
//	 a card."
//
// Skyclave Shadowcat's dies trigger: the creature's toughness is read as it
// last existed (CR 603.10a), counters and anthems included, and the
// Observer's own death counts (its own toughness is 6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dff3cfff-69f9-4afd-af28-b75504d30ffa",
		Name:         "Ulvenwald Observer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					effects.B29LastKnownToughnessOffBattlefield(g, ev.CardID) >= 4
			}, "Ulvenwald Observer — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
