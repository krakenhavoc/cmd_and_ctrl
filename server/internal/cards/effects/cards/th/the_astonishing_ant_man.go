package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Astonishing Ant-Man — Legendary Creature — Human Scientist Hero
// {G}{U}, 1/1 (Marvel Super Heroes, #4455):
//
//	"Whenever you draw a card, put a +1/+1 counter on The Astonishing
//	 Ant-Man.
//	 {2}{G}, {T}, Remove any number of +1/+1 counters from The Astonishing
//	 Ant-Man: Create that many 1/1 green Insect creature tokens."
//
// The draw trigger fires once per card drawn (Chasm Skulker's shape). The
// counter count is announced as the cost is paid (zero is legal) and read
// back from the paid-cost record, since the counters are gone by the time
// the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c7e4e51-099a-4398-a4b7-e2f0ddd36429",
		Name:         "The Astonishing Ant-Man",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "The Astonishing Ant-Man — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, {T}, Remove any number of +1/+1 counters from The Astonishing Ant-Man: Create that many 1/1 green Insect creature tokens.",
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost(), effects.RemoveCountersXFromThis(game.CounterPlusOne, 0)),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := ctx.CountersRemoved()
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Insect"), N: n}.Apply(ctx)
			},
		}},
	})
}
