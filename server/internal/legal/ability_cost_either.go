package legal

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// costBranchRow is one row the activation walk visits: the ability at
// `row`, with — for an either/or cost (#4566) — the branch `branch`
// folded into its cost. `branch` is nil for every other ability.
type costBranchRow struct {
	row    int
	ab     game.ActivatedAbilityShape
	branch *int
}

// expandCostBranches lists the rows abilityMovesForSource walks: each
// ability once, and an ability whose cost has an either/or part once per
// branch, its cost being game.ChosenAbilityCost's fold for that branch.
// The fold is the engine's own, so a move pays exactly what
// ActivateCatalogAbility charges for the branch it names (#544).
func expandCostBranches(abilities []game.ActivatedAbilityShape) []costBranchRow {
	out := make([]costBranchRow, 0, len(abilities))
	for row, ab := range abilities {
		if !ab.Cost.Branched() {
			out = append(out, costBranchRow{row: row, ab: ab})
			continue
		}
		for i := range ab.Cost.Either {
			branch := i
			cost, err := game.ChosenAbilityCost(ab.Cost, &branch)
			if err != nil {
				continue
			}
			folded := ab
			folded.Cost = cost
			out = append(out, costBranchRow{row: row, ab: folded, branch: &branch})
		}
	}
	return out
}
