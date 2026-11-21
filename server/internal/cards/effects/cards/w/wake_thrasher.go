package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wake Thrasher — Creature — Merfolk Soldier {2}{U}, 1/1 (Jumpstart 2022, #5948):
//
//	"Whenever a permanent you control becomes untapped, this creature gets
//	 +1/+1 until end of turn."
//
// One trigger per permanent that untaps (EventUntapCard is raised once per
// permanent, in the untap step as well as by an effect), so each of them
// is a +1/+1 until end of turn, and the Thrasher's own untap counts. It
// only watches permanents under its controller's control as they untap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2d1e991-e13a-43c7-85aa-d53dd14822e6",
		Name:         "Wake Thrasher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventUntapCard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller
			}, "Wake Thrasher — this creature gets +1/+1 until end of turn", effects.ThisGetsUntilEndOfTurn(1, 1, "Wake Thrasher — +1/+1")),
		},
	})
}
