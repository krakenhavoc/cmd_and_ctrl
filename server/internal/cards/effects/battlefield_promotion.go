package effects

// Battlefield Promotion — Instant {1}{W} (Jumpstart Selesnya deck
// request, #4893):
//
//	"Put a +1/+1 counter on target creature. That creature gains first
//	 strike until end of turn. You gain 2 life."
//
// The counter, then first strike until end of turn, on the target if it is
// still legal; then the life. With its only target illegal the spell does
// not resolve at all (CR 608.2b), so no life is gained.
//
// No simplification.
func init() {
	steps := append(plusOneCounterThenKeywordOnTheTarget("first strike", "Battlefield Promotion — first strike until end of turn"),
		GainLife{Amount: 2})
	Register(Spec{
		OracleID:     "303d74d0-84db-4694-bb1a-470cf470bc4f",
		Name:         "Battlefield Promotion",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    ResolveSteps(steps...),
	})
}
