package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shattering Spree — Sorcery {R} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Replicate {R} (When you cast this spell, copy it for each time you
//	 paid its replicate cost. You may choose new targets for the copies.)
//	 Destroy target artifact."
//
// Replicate is the engine's keyword (replicate.go, CR 702.56a) with a mana
// cost: an optional {R} paid any number of times (replicateMana), and
// the cast trigger that copies the spell once per payment, offering each
// copy a new target (CR 707.10c). The copies are created on the stack, not
// cast (the ruling), and the trigger copies the spell from last-known
// information even if it was countered in response. The cap of 20 payments
// on one announcement is the engine's; no table pays for more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "fe725b27-49ef-41bc-91f9-9cd9871ad863",
		Name:          "Shattering Spree",
		Completeness:  CompletenessFull,
		Targets:       TargetPermanent("target artifact", Artifact()),
		OptionalCosts: []game.AdditionalCost{replicateMana("{R}", 20)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		OnResolve:     destroyTheTargetPermanent,
	})
}
