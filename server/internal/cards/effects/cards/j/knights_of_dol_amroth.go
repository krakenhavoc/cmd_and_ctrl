package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knights of Dol Amroth — Creature — Human Knight {3}{U}, 3/3:
//
//	"Whenever you draw your second card each turn, put a +1/+1
//	 counter on this creature."
//
// The trigger is Thopter Fabricator's and Stalwarts of Osgiliath's:
// b41DrewYourSecondCardThisTurn, which reads the turn's draw tally
// after the draw and so fires on the second card and no other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f8b189e-775c-41ad-839a-153a3200c708",
		Name:         "Knights of Dol Amroth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Knights of Dol Amroth — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
