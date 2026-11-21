package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormkeld Prowler — Creature — Human Rogue {1}{U}, 2/1 (Wilds of
// Eldraine, #4471):
//
//	"Whenever you cast a spell with mana value 5 or greater, put two +1/+1
//	 counters on this creature."
//
// Lurking Lizards' cast trigger (WheneverYouCast with ManaValueGE, the
// mana value on the stack with X included, CR 202.3e), with two counters
// for its payout.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9c0e089-92ca-4fbd-9ddd-1802ac3553c9",
		Name:         "Stormkeld Prowler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ManaValueGE(5), "Stormkeld Prowler — put two +1/+1 counters on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
