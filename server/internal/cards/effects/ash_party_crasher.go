package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ash, Party Crasher — Legendary Creature — Human Peasant {R}{W}, 2/2
// (Wilds of Eldraine, #5164):
//
//	"Haste
//	 Celebration — Whenever Ash attacks, if two or more nonland
//	 permanents entered the battlefield under your control this turn,
//	 put a +1/+1 counter on Ash."
//
// Celebration is an intervening if (celebration.go, CR 603.4): checked
// as Ash is declared as an attacker and again as the ability resolves.
// Ash entering this turn counts toward it. The counter goes on Ash only
// while it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1bc701e2-913a-45d5-bc3f-16f87e7e2653",
		Name:            "Ash, Party Crasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacksIfCelebration("Ash, Party Crasher — put a +1/+1 counter on Ash",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
