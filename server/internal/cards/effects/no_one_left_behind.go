package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// No One Left Behind — Sorcery {4}{B} (The Brothers' War, #5482):
//
//	"This spell costs {3} less to cast if it targets a creature card with
//	 mana value 3 or less.
//	 Return target creature card from your graveyard to the battlefield."
//
// The reduction is judged against the announced target when the cost is
// determined (CR 601.2f, after 601.2c). A target with mana value 4 or more
// pays the full {4}{B}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "afe8fd01-f0c7-46a5-9e9c-e8b9599cd19a",
		Name:         "No One Left Behind",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		Purpose:      ForTargets(Reanimates(0)),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a creature card with mana value 3 or less.",
				And(Creature(), ManaValueLE(3))),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return s82Bro10ReturnTargetToBattlefield(false)(ctx.Game, item)
		},
	})
}
