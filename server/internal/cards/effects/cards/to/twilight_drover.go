package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twilight Drover — Creature — Spirit {2}{W}, 1/1 (Tarkir: Dragonstorm Commander):
//
//	"Whenever a creature token leaves the battlefield, put a +1/+1
//	 counter on this creature.
//	 {2}{W}, Remove a +1/+1 counter from this creature: Create two 1/1
//	 white Spirit creature tokens with flying."
//
// The trigger reads the leaving permanent as it last existed (CR 603.10a):
// a creature, by its last-known types, and a token. Any controller, any
// destination: it watches every EventLTB. The Drover's own death is not
// "a creature token" unless it is one. The activated ability's counter is
// a cost (CR 118), so it can't be responded to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0d60856-80f9-4c64-9768-28a0a690ecb1",
		Name:         "Twilight Drover",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.IsToken(c) && effects.LeftAsType(ev, c, "creature")
			}, "Twilight Drover — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, Remove a +1/+1 counter from this creature: Create two 1/1 white Spirit creature tokens with flying.",
			Purpose: game.Purpose{Tokens: 2, CreatureTokens: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
