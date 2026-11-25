package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chira, All In — {3}{R} Legendary Creature — Human Rogue 3/3:
//
//	"Whenever you discard one or more cards at random, select one or more
//	 targets at random, choose one or more modes at random, flip one or more
//	 coins, or roll one or more dice, put a +1/+1 counter on Chira and
//	 create a tapped Treasure token."
//
// One printed trigger, once per instruction (#6101): a random discard, a
// random target selection and a random mode choice each announce
// themselves once per instruction (game/random_events.go); coins and dice
// announce one event per coin or die, and only the first of an
// instruction's batch (Seq == BatchSeq) triggers Chira, as "whenever you
// roll one or more dice" does. "You" is the player who discards, selects,
// chooses, flips or rolls, whoever's effect told them to.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6552921-f949-4f3c-b75d-718c106d712b",
		Name:         "Chira, All In",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{
				game.EventDiscardedAtRandom, game.EventTargetsSelectedAtRandom, game.EventModesChosenAtRandom,
				game.EventFlipCoin, game.EventRollDie,
			}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				switch ev.Kind {
				case game.EventFlipCoin, game.EventRollDie:
					return ev.Seq == ev.BatchSeq
				}
				return true
			}, "Chira, All In — put a +1/+1 counter on Chira and create a tapped Treasure token",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1},
					effects.CreateTokenAdvanced{Spec: effects.Token(effects.TreasureToken()).EntersTapped(), N: 1})),
		},
	})
}
