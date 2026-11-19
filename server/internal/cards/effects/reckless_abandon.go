package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Abandon — Sorcery {R} (Urza's Destiny, #5621):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Reckless Abandon deals 4 damage to any target."
//
// Improvised Club's shape with a creature only: the sacrifice is an
// additional cost paid as the spell is cast (CR 118.8), so it is gone
// even if the spell is countered, and the spell can't be cast without a
// creature to sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "f24d5187-0f80-4528-bf44-03ae146f3551",
		Name:           "Reckless Abandon",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetAny(),
		Purpose:        WithRiders(ForTargets(DamageToTarget(0, 4)), game.Purpose{Answers: game.AnswerSacOutlet}),
		OnResolve:      damageToFirstTarget(4),
	})
}
