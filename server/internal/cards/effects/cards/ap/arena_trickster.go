package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arena Trickster — Creature — Human Shaman {3}{R}, 3/3 (Theros Beyond Death,
// #5514):
//
//	"Whenever you cast your first spell during each opponent's turn, put a
//	 +1/+1 counter on this creature."
//
// Stinging Lionfish's trigger (the first spell its controller casts in a
// turn that is an opponent's, by the cast tally), with the counter put on
// the Trickster as it resolves; if it has left by then nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69c27223-d676-4916-899b-297dc4ed2d59",
		Name:         "Arena Trickster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B22FirstSpellOnAnOpponentsTurn(ev, source, g)
			}, "Arena Trickster — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
