package game

import (
	"fmt"

	"github.com/google/uuid"
)

// ability_cost_either.go — #4566 (ADR 0100's 2026-10-30 amendment): an
// either/or part of an activated ability's cost. Bullseye, Death
// Dealer prints "{3}, {T}, Sacrifice an artifact or discard a nonland
// card: Bullseye deals 2 damage to any target."
//
// The cast path has had this since ADR 0100 §2 (AdditionalCost.Either,
// CastSpellParams.CostBranch), and this is the same decision one
// payer over: CR 602.2b runs CR 601.2b–h for an ability, so the
// activator announces which branch they will pay with the rest of the
// announcement, the branch is paid at CR 601.2h with everything else,
// and a branch that can't be paid can't be chosen (CR 118.3). One
// ability with a choice, not two abilities: the card prints one.
//
// The design is a fold. ChosenAbilityCost returns the cost with the
// announced branch turned into the ordinary sacrifice or discard clause
// and Either cleared, and the activation, the enumerator and the view
// hand THAT to the code they already run. The validator and the payer
// therefore never learn that branches exist, which is the point ADR
// 0073's payment plan made for casts.
//
// A branch is plain data (AbilityCostBranch), not an AbilityCost: a
// sacrifice or discard clause carries a predicate func, and a new field
// of that type would be a new route from Game to a closure (the ADR 0041
// ratchet). The clause is built from the data when the branch is folded.

// AbilityCostBranchKind names what one branch of an either/or activation
// cost pays.
type AbilityCostBranchKind string

const (
	// AbilityCostBranchSacrifice sacrifices Count permanents the
	// activator controls that match Of.
	AbilityCostBranchSacrifice AbilityCostBranchKind = "sacrifice"
	// AbilityCostBranchDiscard discards Count cards from the activator's
	// hand: nonland cards only when Nonland is set.
	AbilityCostBranchDiscard AbilityCostBranchKind = "discard"
)

// AbilityCostBranch is one way to pay an either/or part of an activated
// ability's cost (AbilityCost.Either). Plain data; ChosenAbilityCost
// builds the clause it stands for.
type AbilityCostBranch struct {
	Kind AbilityCostBranchKind
	// Count is how many permanents or cards the branch takes.
	Count int
	// Label is the printed object of the verb: "an artifact", "a
	// nonland card".
	Label string
	// Of narrows a sacrifice ("an artifact"). The zero value is any
	// permanent. Unused by a discard.
	Of PermanentQuery
	// Nonland narrows a discard to nonland cards. Unused by a
	// sacrifice.
	Nonland bool
}

// ErrAbilityCostBranch is returned when an activation names no branch
// of an either/or cost, a branch the cost does not have, or a branch on
// an ability whose cost has none (#4566). The ErrCostBranch posture: no
// branch is ever guessed.
var ErrAbilityCostBranch = fmt.Errorf("%w: bad either/or activation cost branch", ErrInvalidParam)

// Branched reports whether the cost has an either/or part.
func (c AbilityCost) Branched() bool { return len(c.Either) > 0 }

// ChosenAbilityCost is the cost one activation pays: `c` itself for a
// cost with no branches, or — for an either/or cost — `c` with branch
// `*branch` built into its sacrifice or discard clause and Either
// cleared. The branch is REQUIRED on a branched cost and refused on any
// other (ErrAbilityCostBranch).
//
// effects.Register keeps the rest of a branched cost free of both
// clauses, so the fold never replaces a printed one. Pure, so the
// activation, the enumerator and the view ask it the same question.
func ChosenAbilityCost(c AbilityCost, branch *int) (AbilityCost, error) {
	if !c.Branched() {
		if branch != nil {
			return c, ErrAbilityCostBranch
		}
		return c, nil
	}
	if branch == nil || *branch < 0 || *branch >= len(c.Either) {
		return c, ErrAbilityCostBranch
	}
	b := c.Either[*branch]
	out := c
	out.Either = nil
	switch b.Kind {
	case AbilityCostBranchSacrifice:
		out.SacrificeOther = b.sacrificeClause()
	case AbilityCostBranchDiscard:
		out.DiscardCards = b.discardClause()
	default:
		return c, ErrAbilityCostBranch
	}
	return out, nil
}

// count is the branch's Count, with zero reading as one.
func (b AbilityCostBranch) count() int {
	if b.Count < 1 {
		return 1
	}
	return b.Count
}

// sacrificeClause is the "Sacrifice <Label>" clause a sacrifice branch
// stands for: Count permanents on the battlefield matching Of, by their
// effective characteristics.
func (b AbilityCostBranch) sacrificeClause() *TargetSpec {
	q := b.Of
	n := b.count()
	return &TargetSpec{
		Mode:  "permanent",
		Label: b.Label,
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(g *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return q.matchesLocked(g, &c)
		},
		Min: n, Max: n,
	}
}

// discardClause is the "Discard <Label>" clause a discard branch stands
// for.
func (b AbilityCostBranch) discardClause() *DiscardCost {
	d := &DiscardCost{N: b.count(), Label: b.Label}
	if b.Nonland {
		d.Match = func(c Card) bool { return !c.IsLand() }
	}
	return d
}

// BranchLabel is how a branch reads on its own: "Sacrifice an
// artifact", "Discard a nonland card". The view's branch picker names a
// branch with it.
func (b AbilityCostBranch) BranchLabel() string {
	switch b.Kind {
	case AbilityCostBranchSacrifice:
		return "Sacrifice " + b.Label
	case AbilityCostBranchDiscard:
		return "Discard " + b.Label
	}
	return b.Label
}
