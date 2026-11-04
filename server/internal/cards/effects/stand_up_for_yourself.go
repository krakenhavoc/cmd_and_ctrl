package effects

// Stand Up for Yourself — Instant {2}{W} (Secrets of Strixhaven, #4456):
//
//	"Destroy target creature with power 3 or greater."
//
// Bring Down's shape: the pick is a creature with power 3 or greater
// (PowerGE), re-checked at resolution (CR 608.2b), so a creature shrunk
// in response is an illegal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00e94b75-ad31-4e1f-b3be-eed7cd3e4d76",
		Name:         "Stand Up for Yourself",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with power 3 or greater", PowerGE(3)),
		OnResolve:    destroyTheTargetPermanent,
	})
}
