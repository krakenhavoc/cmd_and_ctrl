package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// or_cost.go — #4566: Register's guard for an either/or part of an
// activated ability's cost (OrCost, game.AbilityCost.Either).
//
// game.ChosenAbilityCost builds the announced branch into the cost's
// sacrifice or discard clause, so the rest of the cost may carry neither
// (the fold would replace the printed one), and a branch must be one of
// the two kinds with a count and the printed words. Card-file mistakes,
// refused at boot.

// checkOrCost panics on a malformed either/or cost on ability `i`.
func checkOrCost(name string, i int, cost game.AbilityCost) {
	if !cost.Branched() {
		return
	}
	where := fmt.Sprintf("ability %d", i)
	if len(cost.Either) < 2 {
		panic(fmt.Sprintf("effects.Register: %q %s declares an either/or cost with %d branch(es) — build it with OrCost over two or more", name, where, len(cost.Either)))
	}
	if cost.SacrificeOther != nil || cost.DiscardCards != nil {
		panic(fmt.Sprintf("effects.Register: %q %s has an either/or cost beside its own sacrifice or discard clause — the branch would replace it (game.ChosenAbilityCost)", name, where))
	}
	for b, branch := range cost.Either {
		switch {
		case branch.Kind != game.AbilityCostBranchSacrifice && branch.Kind != game.AbilityCostBranchDiscard:
			panic(fmt.Sprintf("effects.Register: %q %s branch %d has kind %q — a branch sacrifices or discards (SacrificeBranch, DiscardBranch)", name, where, b, branch.Kind))
		case branch.Count < 1 || branch.Label == "":
			panic(fmt.Sprintf("effects.Register: %q %s branch %d needs a count of at least one and its printed words", name, where, b))
		case branch.Kind == game.AbilityCostBranchDiscard && !branch.Of.Equal(game.PermanentQuery{}):
			panic(fmt.Sprintf("effects.Register: %q %s branch %d discards but names a permanent query", name, where, b))
		case branch.Kind == game.AbilityCostBranchSacrifice && branch.Nonland:
			panic(fmt.Sprintf("effects.Register: %q %s branch %d sacrifices but says nonland — a sacrifice is narrowed by its query", name, where, b))
		}
	}
}
