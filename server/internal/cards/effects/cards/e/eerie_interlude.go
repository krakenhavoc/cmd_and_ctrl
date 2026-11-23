package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eerie Interlude — Instant {2}{W} (EDHREC top 1,000, #3654):
//
//	"Exile any number of target creatures you control. Return those
//	 cards to the battlefield under their owner's control at the
//	 beginning of the next end step."
//
// Waterbender's Restoration's shape with a free count: every still-legal
// target is exiled as one event and the cards that reached exile come
// back together at the next end step as new objects (CR 400.7). A target
// that became illegal in response stays where it is (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0634091a-a74c-4cea-b6d1-7324a725554a",
		Name:         "Eerie Interlude",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("any number of target creatures you control", effects.YouControl()).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTargetsThenScheduleReturn(ctx, "Eerie Interlude — return the exiled creatures")
		},
	})
}
