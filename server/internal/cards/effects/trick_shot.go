package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trick Shot — Instant {4}{R} (Outlaws of Thunder Junction, #4468):
//
//	"Trick Shot deals 6 damage to target creature and 2 damage to up to
//	 one other target creature token."
//
// Conduct Electricity's body (damageEachClauseTarget) with the second
// clause Distinct: "other" means the token cannot be the creature already
// chosen for the 6 (CR 115.3). The second clause is optional. A clause
// whose target has become illegal is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "843c8485-21b8-45ef-9a0c-4967f07278bb",
		Name:         "Trick Shot",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("up to one other target creature token", IsTokenPredicate()).WithCount(0, 1)),
		),
		Purpose:   game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 6), DamageToTarget(1, 2)), Answers: game.AnswerRemove},
		OnResolve: damageEachClauseTarget(6, 2),
	})
}
