package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deep Goblin Skulltaker — Creature — Goblin Warrior {2}{B}, 2/2 (The
// Lost Caverns of Ixalan, #4470):
//
//	"Menace
//	 At the beginning of your end step, if you descended this turn, put a
//	 +1/+1 counter on this creature."
//
// Menace is the printed keyword (CR 702.111b). "You descended this turn"
// means a permanent card was put into your graveyard from anywhere this
// turn (CR 700.11), so a token dying does not count. The "if" is an
// intervening if (CR 603.4), checked as your end step begins and again as
// the trigger resolves (AtYourEndStepIfDescended). The counter goes on
// the Skulltaker only while it is still this object on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7f56bbb-961a-4641-9446-7292ef188b1d",
		Name:            "Deep Goblin Skulltaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfDescended("Deep Goblin Skulltaker — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: "+1/+1", N: 1})),
		},
	})
}
