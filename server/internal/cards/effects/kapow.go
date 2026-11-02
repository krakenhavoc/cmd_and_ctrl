package effects

// Kapow! — Sorcery {2}{G} (Marvel's Spider-Man, #4460):
//
//	"Put a +1/+1 counter on target creature you control. It fights target
//	 creature an opponent controls."
//
// Savage Stomp's two steps without its Dinosaur discount: the counter lands
// first, so the fight uses the grown power, and the fight is skipped if
// either target is no longer legal (CR 608.2b) while the counter still
// lands on a legal creature of yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6eabbff-cc0c-4000-8feb-4826e1b6bc66",
		Name:         "Kapow!",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls())),
		OnResolve: ResolveSteps(Instruction(counterOnFirstCreatureTarget), Instruction(fightTargetClauses)),
	})
}
