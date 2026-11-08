package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// This Town Ain't Big Enough — Instant {4}{U} (Outlaws of Thunder Junction,
// #4468):
//
//	"This spell costs {3} less to cast if it targets a permanent you
//	 control.
//	 Return up to two target nonland permanents to their owners' hands."
//
// The discount is a self cost modifier that reads the announced targets
// (CostsLessIfItTargets, as Savage Stomp's does): one of the chosen
// permanents being yours is enough, and the clause only picks nonland
// permanents, so that is the "permanent you control" it can read. The
// bounce is Aether Gale's (b35BounceChosenPermanents): every chosen
// permanent still legal as the spell resolves returns at once (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2550c2ae-834d-46e7-9ab4-d7bf7bfa6a78",
		Name:         "This Town Ain't Big Enough",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("up to two target nonland permanents", Nonland()).WithCount(0, 2),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a permanent you control.", YouControl()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b35BounceChosenPermanents(ctx)
		},
	})
}
