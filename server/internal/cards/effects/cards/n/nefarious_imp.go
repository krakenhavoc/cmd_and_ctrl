package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nefarious Imp — {2}{B} Creature — Imp 2/1
//
//	"Flying
//	 Whenever one or more permanents you control leave the battlefield, scry 1."
//
// One trigger per batch of departures (CR 603.2c). Control is read as the
// permanent last had it (CR 603.10a), so the Imp's own exit counts.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c412191f-46d7-4520-a76c-1f70a2e92ab6",
		Name:            "Nefarious Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftUnderControlOf(ev, c) == source.Controller
			}, "Nefarious Imp — scry 1", effects.Do(effects.Scry{N: 1}))),
		},
	})
}
