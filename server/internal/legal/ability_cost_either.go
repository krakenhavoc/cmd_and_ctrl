package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// costBranchRow is one row the activation walk visits: the ability at
// `row`, with — for an either/or cost (#4566) — the branch `branch`
// folded into its cost. `branch` is nil for every other ability.
//
// `unattach` is the Equipment an "Unattach an Equipment from <this>"
// cost names on this row (#4799, expandUnattachPicks), nil for every
// other ability.
//
// `blight` is the creature a "Blight N" cost names on this row (#4684,
// expandBlightPicks), nil for every other ability.
type costBranchRow struct {
	row      int
	ab       game.ActivatedAbilityShape
	branch   *int
	unattach []uuid.UUID
	blight   []uuid.UUID
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

// manaCostBranchRow is one row the mana walk visits: the mana ability at
// `row`, with — for an either/or cost (#5397, forage) — the branch
// `branch` folded in. `branch` is nil for every other mana ability.
type manaCostBranchRow struct {
	row    int
	ab     game.ManaAbilityShape
	branch *int
}

// expandManaCostBranches is expandCostBranches for mana abilities: each
// ability once, and one whose cost has an either/or part once per
// branch, folded by game.ChosenManaAbilityShape — the fold
// ActivateManaAbility pays (#544).
func expandManaCostBranches(abilities []game.ManaAbilityShape) []manaCostBranchRow {
	out := make([]manaCostBranchRow, 0, len(abilities))
	for row, ab := range abilities {
		if !ab.Branched() {
			out = append(out, manaCostBranchRow{row: row, ab: ab})
			continue
		}
		for i := range ab.Either {
			branch := i
			folded, err := game.ChosenManaAbilityShape(ab, &branch)
			if err != nil {
				continue
			}
			out = append(out, manaCostBranchRow{row: row, ab: folded, branch: &branch})
		}
	}
	return out
}
