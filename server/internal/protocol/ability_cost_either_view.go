package protocol

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// AbilityCostBranchView is one way to pay an either/or part of an
// activated ability's cost (#4566): the branch's words, whether the
// activator can pay it now (CR 118.3), and the picker it needs — the
// same fields an ability printing that clause alone carries, so the
// client runs the picker it already has.
type AbilityCostBranchView struct {
	// Label is the branch as it reads alone: "Sacrifice an artifact",
	// "Discard a nonland card".
	Label string `json:"label"`
	// Payable is whether the activator has enough to pay this branch
	// right now. The client greys a branch that is not.
	Payable bool `json:"payable,omitempty"`
	// SacrificeLabel and SacrificeOptions are a sacrifice branch's
	// words and the permanents that could pay it (the battlefield, so
	// public).
	SacrificeLabel   string            `json:"sacrifice_label,omitempty"`
	SacrificeOptions *LegalTargetsView `json:"sacrifice_options,omitempty"`
	// DiscardCostN, DiscardCostLabel and DiscardCostOptions are a discard
	// branch's count, words and the matching cards in the activator's
	// hand. The options are the controller's alone (#1369).
	DiscardCostN       int      `json:"discard_cost_n,omitempty"`
	DiscardCostLabel   string   `json:"discard_cost_label,omitempty"`
	DiscardCostOptions []string `json:"discard_cost_options,omitempty"`
}

// abilityCostBranchViews projects the branches of an either/or cost,
// each from the cost game.ChosenAbilityCost folds for it, through the
// candidate walks the activation validates against, so a branch the
// view calls payable is one the engine accepts (#544). Nil for a cost
// with no branches. Caller must hold g.mu.
func abilityCostBranchViews(g *game.Game, caster, sourceID uuid.UUID, cost game.AbilityCost) []AbilityCostBranchView {
	if !cost.Branched() {
		return nil
	}
	out := make([]AbilityCostBranchView, 0, len(cost.Either))
	for i := range cost.Either {
		branch := i
		folded, err := game.ChosenAbilityCost(cost, &branch)
		if err != nil {
			continue
		}
		b := AbilityCostBranchView{Label: cost.Either[i].BranchLabel(), Payable: true}
		if sc := folded.SacrificeOther; sc != nil {
			b.SacrificeLabel = sc.Label
			b.SacrificeOptions = sacrificeCostOptions(g, caster, sc, sourceID,
				folded.SacrificeSelf || folded.ReturnSelf || folded.BottomSelf)
			b.Payable = len(b.SacrificeOptions.Cards) >= b.SacrificeOptions.Min
		}
		if dc := folded.DiscardCards; dc != nil {
			b.DiscardCostN = dc.N
			b.DiscardCostLabel = dc.Label
			b.DiscardCostOptions = cardIDStrings(g.DiscardCostOptionsForEffect(caster, sourceID, dc))
			b.Payable = b.Payable && len(b.DiscardCostOptions) >= dc.N
		}
		out = append(out, b)
	}
	return out
}

// publicCostBranches is the half of an either/or cost's branches every
// viewer may read (#1369): a discard branch's hand cards are cleared,
// and so is its verdict, which would say whether the controller holds a
// matching card. `private` reports whether anything was cleared.
func publicCostBranches(in []AbilityCostBranchView) (out []AbilityCostBranchView, private bool) {
	if len(in) == 0 {
		return in, false
	}
	out = make([]AbilityCostBranchView, len(in))
	for i, b := range in {
		if b.DiscardCostN > 0 || len(b.DiscardCostOptions) > 0 {
			private = private || len(b.DiscardCostOptions) > 0 || b.Payable
			b.DiscardCostOptions = nil
			b.Payable = false
		}
		out[i] = b
	}
	return out, private
}
