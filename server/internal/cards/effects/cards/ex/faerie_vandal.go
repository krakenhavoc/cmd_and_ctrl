package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Vandal — Creature — Faerie Rogue {1}{U}, 1/2:
//
//	"Flash
//	 Flying
//	 Whenever you draw your second card each turn, put a +1/+1 counter
//	 on this creature."
//
// Atlantean Cavalry's trigger word for word (b41DrewYourSecondCardThisTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b872f147-00eb-4cf8-af8c-4144356b2089",
		Name:            "Faerie Vandal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Faerie Vandal — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
