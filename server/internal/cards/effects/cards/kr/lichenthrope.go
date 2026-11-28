package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lichenthrope — Creature — Plant Fungus {3}{G}{G}, 5/5 (Visions, #5631):
//
//	"If damage would be dealt to this creature, put that many -1/-1
//	 counters on it instead.
//	 At the beginning of your upkeep, remove a -1/-1 counter from this
//	 creature."
//
// The first sentence is Phytohydra's replacement with -1/-1 counters
// (DamageToCountersInstead): a replacement, not a prevention, so it applies
// to damage that can't be prevented and the damage is not dealt. The upkeep
// trigger removes one -1/-1 counter if it has any.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee7f698d-55c5-4a7a-803d-304febe6a758",
		Name:         "Lichenthrope",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.DamageToCountersInstead("Lichenthrope — put that many -1/-1 counters on it instead of dealing damage to it", game.CounterMinusOne),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Lichenthrope — remove a -1/-1 counter from it", effects.RemoveACounterFromThis(game.CounterMinusOne)),
		},
	})
}
