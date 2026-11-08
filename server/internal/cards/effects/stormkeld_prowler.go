package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d9c0e089-92ca-4fbd-9ddd-1802ac3553c9",
		Name:         "Stormkeld Prowler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(5), "Stormkeld Prowler — put two +1/+1 counters on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
