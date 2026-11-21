package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brineborn Cutthroat — "Flash. Whenever you cast a spell during an
// opponent's turn, put a +1/+1 counter on this creature."
//
// The flash payoff: every instant-speed play on someone else's turn
// grows it. "During an opponent's turn" is read off the turn cursor
// at trigger time — the active seat's player is not the caster.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "916cb70f-3b06-48ed-972d-75f805aa0892",
		Name:            "Brineborn Cutthroat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				return !effects.IsActivePlayer(g, source.Controller)
			}, "Brineborn Cutthroat — +1/+1 counter", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
