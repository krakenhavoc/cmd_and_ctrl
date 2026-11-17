package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elvish Fury — Instant {G} (Tempest, #5628):
//
//	"Buyback {4} (You may pay an additional {4} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Target creature gets +2/+2 until end of turn."
//
// Artful Maneuver's pump (boostTheTargetUntilEOT at +2/+2, ending in the
// cleanup step, CR 514.2) with Capsize's buyback (CR 702.27a). A target
// that left in response means the spell doesn't resolve (CR 608.2b) and
// goes to the graveyard, buyback paid or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "557d2af0-04c5-461d-826b-7208c4535dc2",
		Name:          "Elvish Fury",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Buyback("{4}")},
		OnResolve:     boostTheTargetUntilEOT(2, 2, "Elvish Fury — +2/+2"),
	})
}
