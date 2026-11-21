package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestial Unicorn — Creature — Unicorn {2}{W}, 3/2 (Adventures in the
// Forgotten Realms, S82):
//
//	"Whenever you gain life, put a +1/+1 counter on this creature."
//
// Ajani's Pridemate's trigger: once per life-gain event, not once per
// point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be030639-cbc9-416b-b0ca-1a76439791d8",
		Name:         "Celestial Unicorn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LifeGainPaying(
				effects.WheneverYouGainLife("Celestial Unicorn — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
		},
	})
}
