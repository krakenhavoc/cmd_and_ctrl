package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Ambush — Sorcery {1}{R} (Portal Three Kingdoms, #5622):
//
//	"Fire Ambush deals 3 damage to any target."
//
// Volcanic Hammer's body (damageToFirstTarget). A target that became
// illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50463946-1ce3-4ff0-ad68-2fb87adbe2fd",
		Name:         "Fire Ambush",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(3),
	})
}
