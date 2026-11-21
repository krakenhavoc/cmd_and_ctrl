package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leela, Sevateem Warrior — Legendary Creature — Human Warrior {3}{G}, 3/3
// (Doctor Who):
//
//	"Whenever an opponent draws a card except the first one they draw in each
//	 of their draw steps, put a +1/+1 counter on Leela.
//	 Doctor's companion"
//
// Orcish Bowmasters' draw condition: only the first completed draw of an
// opponent's own draw step is exempt. Doctor's companion is a
// deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ae972ea-d5d8-4064-9d91-5b2040d7dcbf",
		Name:         "Leela, Sevateem Warrior",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ByAnOpponent(ev, source, ch, g) && !g.FirstCompletedDrawInOwnDrawStepForEffect(ev.Actor)
			}, "Leela — put a +1/+1 counter on Leela", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
