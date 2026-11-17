package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seething Anger — Sorcery {R} (Stronghold, #5627):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Target creature gets +3/+0 until end of turn."
//
// Elvish Fury's shape: the shared boostTheTargetUntilEOT body at +3/+0,
// with Capsize's buyback (CR 702.27a). The boost ends in the cleanup step
// (CR 514.2). A target that became illegal in response means the spell
// doesn't resolve (CR 608.2b), so it goes to the graveyard, buyback paid
// or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "2effcd1c-5833-4e0e-bb67-2d2ffb22a014",
		Name:          "Seething Anger",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(BenefitUntilEOT(0, 3, 0)),
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve:     boostTheTargetUntilEOT(3, 0, "Seething Anger — +3/+0"),
	})
}
