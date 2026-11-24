package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Drake — Creature — Drake {2}{U}, 4/3 (Modern Horizons 2, #5501):
//
//	"Flying
//	 When this creature enters, return two target lands you control to
//	 their owner's hand."
//
// Exactly two targets: with fewer than two lands the trigger is never put
// on the stack (CR 603.3d). Each is returned if still legal as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "44d1029f-604e-44d6-89f6-bd18344f5d79",
		Name:            "Sea Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Sea Drake — return two target lands you control to their owner's hand",
				effects.BounceEachLegalTarget),
				effects.TargetPermanent("two target lands you control", effects.Land(), effects.YouControl()).WithCount(2, 2)),
		},
	})
}
