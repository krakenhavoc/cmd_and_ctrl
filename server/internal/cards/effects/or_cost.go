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
// the two kinds with a count and the printed words. #5184 adds forage's
// graveyard exile, a third kind only Forage() may use, and holds the
// Forage mark to forage's two branches. Card-file mistakes, refused at
// boot.

// checkOrCost panics on a malformed either/or cost on ability `i`.
func checkOrCost(name string, i int, cost game.AbilityCost) {
	where := fmt.Sprintf("ability %d", i)
	if cost.Forage && !isForageBranches(cost.Either) {
		// #5184: the mark names the forage branches and nothing else.
		panic(fmt.Sprintf("effects.Register: %q %s marks a cost Forage over branches that are not forage's — build it with Forage()", name, where))
	}
	if !cost.Branched() {
		return
	}
	if len(cost.Either) < 2 {
		panic(fmt.Sprintf("effects.Register: %q %s declares an either/or cost with %d branch(es) — build it with OrCost over two or more", name, where, len(cost.Either)))
	}
	if cost.SacrificeOther != nil || cost.DiscardCards != nil || cost.ExileCards != nil {
		panic(fmt.Sprintf("effects.Register: %q %s has an either/or cost beside its own sacrifice, discard or exile clause — the branch would replace it (game.ChosenAbilityCost)", name, where))
	}
	for b, branch := range cost.Either {
		if branch.Kind == game.AbilityCostBranchMana {
			if cost.Mana != "" || branch.Mana == "" {
				panic(fmt.Sprintf("effects.Register: %q %s branch %d pays mana beside the cost's own mana, or names none (ManaBranch)", name, where, b))
			}
			continue
		}
		switch {
		case branch.Kind == game.AbilityCostBranchExileGraveyard && !cost.Forage:
			panic(fmt.Sprintf("effects.Register: %q %s branch %d exiles from a graveyard — only forage's branch does (Forage)", name, where, b))
		case branch.Kind != game.AbilityCostBranchSacrifice && branch.Kind != game.AbilityCostBranchDiscard && branch.Kind != game.AbilityCostBranchExileGraveyard:
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

// isForageBranches reports whether `branches` are
// game.ForageAbilityBranches, in order.
func isForageBranches(branches []game.AbilityCostBranch) bool {
	want := game.ForageAbilityBranches()
	if len(branches) != len(want) {
		return false
	}
	for i, b := range branches {
		w := want[i]
		if b.Kind != w.Kind || b.Count != w.Count || b.Label != w.Label || b.Nonland != w.Nonland || !b.Of.Equal(w.Of) {
			return false
		}
	}
	return true
}
