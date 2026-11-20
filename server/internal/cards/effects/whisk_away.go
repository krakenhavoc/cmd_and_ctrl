package effects

// Whisk Away — Instant {2}{U} (Fate Reforged,
// #5552):
//
//	"Put target attacking or blocking creature on top of its owner's
//	 library."
//
// Time Ebb's tuck with the clause narrowed to an attacking or blocking
// creature, judged when it is targeted and again as the spell resolves (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac141fe0-328a-4d9b-8048-b0c47f399587",
		Name:         "Whisk Away",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve:    putTargetOnTopOfOwnersLibrary,
	})
}
