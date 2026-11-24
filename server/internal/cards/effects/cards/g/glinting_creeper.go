package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glinting Creeper — Creature — Plant {4}{G}, 0/0 (Modern Horizons 2,
// #5501):
//
//	"Converge — This creature enters with two +1/+1 counters on it for each
//	 color of mana spent to cast it.
//	 This creature can't be blocked by creatures with power 2 or less."
//
// Converge is Wildgrowth Archaic's count of the colours spent on the
// resolving spell (CR 614.1c), doubled; a Creeper that was not cast enters
// as a 0/0 and dies. The evasion is Brushfire Elemental's block rule over
// the blocker's current power.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "0e252f94-7731-45e1-bdbb-1bd11441577e",
		Name:                "Glinting Creeper",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind:  game.CounterPlusOne,
			Count: func(cast game.CastCounts) int { return 2 * cast.ColorsSpent },
		}},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
