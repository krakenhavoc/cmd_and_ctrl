package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Searing Touch — Instant {R} (Tempest, #5628):
//
//	"Buyback {4} (You may pay an additional {4} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Searing Touch deals 1 damage to any target."
//
// Shock's target and the shared damageToFirstTarget body at 1, with
// Capsize's buyback (CR 702.27a). A target that became illegal in
// response means the spell doesn't resolve (CR 608.2b), so it goes to
// the graveyard even when the buyback was paid, as Capsize's does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "1cd9a3ba-a115-4e59-983d-605e4435f3a8",
		Name:          "Searing Touch",
		Completeness:  CompletenessFull,
		Targets:       TargetAny(),
		Purpose:       game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 1)), Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{Buyback("{4}")},
		OnResolve:     damageToFirstTarget(1),
	})
}
