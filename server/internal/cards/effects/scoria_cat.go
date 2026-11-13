package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scoria Cat — Creature — Cat {3}{R}{R}, 3/3 (Prophecy, #5794):
//
//	"This creature gets +3/+3 as long as you control no untapped lands."
//
// A layer 7c self pump (CR 613.4c) while its controller controls no
// untapped lands (no_untapped_lands.go). The condition is read live
// (CR 611.3a): the Cat is 6/6 the moment the last land taps and 3/3 the
// moment one untaps or an untapped land enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88ef3413-21bc-428a-9803-d275e437c4bd",
		Name:         "Scoria Cat",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{SelfPumpWhileNoUntappedLands(3, 3)},
	})
}
