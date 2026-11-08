package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Child of the Volcano — Creature — Elemental {3}{R}, 3/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Trample
//	 At the beginning of your end step, if you descended this turn, put a
//	 +1/+1 counter on this creature."
//
// Deep Goblin Skulltaker's trigger, with trample in place of menace.
// "You descended this turn" means a permanent card was put into your
// graveyard from anywhere this turn (CR 700.11), so a token dying does
// not count. The "if" is an intervening if (CR 603.4), checked as your
// end step begins and again as the trigger resolves
// (AtYourEndStepIfDescended). The counter goes on this creature only
// while it is still this object on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "701c8624-70d8-49df-84b1-d7b2a8e83552",
		Name:            "Child of the Volcano",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfDescended("Child of the Volcano — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
