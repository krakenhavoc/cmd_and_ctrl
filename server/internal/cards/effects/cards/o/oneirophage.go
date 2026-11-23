package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oneirophage — Creature — Squid Illusion {3}{U}, 1/2 (Jumpstart, #4892):
//
//	"Flying
//	 Whenever you draw a card, put a +1/+1 counter on this creature."
//
// The trigger fires once per card you draw, on any player's turn; an
// opponent's draw does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "48c97aeb-6591-4175-a24b-d3488bf3ab0b",
		Name:            "Oneirophage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "Oneirophage — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
