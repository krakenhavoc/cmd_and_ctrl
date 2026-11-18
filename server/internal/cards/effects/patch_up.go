package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patch Up — Sorcery {2}{W} (Streets of New Capenna):
//
//	"Return up to three target creature cards with total mana value 3 or
//	 less from your graveyard to the battlefield."
//
// Scout for Survivors without the counters: one clause under the total set
// rule (TotalManaValueAtMost), judged as the spell is cast (CR 601.2c) and
// again over the survivors as it resolves (CR 608.2b); every card still in
// the graveyard enters together (CR 603.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34a7026f-c6ca-4d89-8364-fd60fb613634",
		Name:         "Patch Up",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("up to three target creature cards with total mana value 3 or less from your graveyard",
			YouOwn(), Creature()).WithCount(0, 3).WithTotal(TotalManaValueAtMost(3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{Targets: legalTargetCardIDs(ctx)}.Apply(ctx)
		},
	})
}
