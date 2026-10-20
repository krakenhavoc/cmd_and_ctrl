package game

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

// tap_cost_chain.go — a cast that owes more than one tap-permanents cost
// (#4177, ADR 0073's 2026-10-19 amendment). Hama, the Bloodbender's cast
// is priced by waterbending {N}, and the spell may ALSO have convoke or
// improvise (its own, a permanent's grant, Archway of Innovation's
// promise), or its own waterbend additional cost (Waterbender's
// Restoration), or a chosen optional waterbend (Secret of Bloodbending).
// Each is a cost component with its own pool and its own share of the
// bill (CR 701.67b: a waterbend's taps pay only its own generic; CR
// 702.51a: convoke pays for the whole cost), and a permanent taps once.
//
// The wire is unchanged: a cast still names ONE list of tapped permanents
// (CastSpellParams.TapIDs). The costs are linked through
// TapPermanentsCost.Also (references, rebuilt by tapCostFromRef), and the engine shares the list between them
// (splitTapPayersLocked), so the player chooses what to tap and the
// server works out which cost each tap pays.

// ConvokeTapCost is the convoke cost a chained link is rebuilt from. The
// effects package sets it at init, as it does ImproviseTapCost.
var ConvokeTapCost *TapPermanentsCost

// tapCostFromRef rebuilds one chained link, or nil when its constructor
// is not installed.
func tapCostFromRef(r TapCostRef) *TapPermanentsCost {
	var out *TapPermanentsCost
	switch r.Key {
	case "convoke":
		out = ConvokeTapCost
	case "improvise":
		out = ImproviseTapCost
	case "waterbend":
		if WaterbendTapCost != nil {
			out = WaterbendTapCost(r.Extra)
		}
	}
	if out.Empty() {
		return nil
	}
	cp := *out
	cp.Also = nil
	cp.InCost = r.InCost
	return &cp
}

// HasChain reports whether the cost carries chained tap costs. Nil-safe.
func (c *TapPermanentsCost) HasChain() bool { return c != nil && len(c.Also) > 0 }

// links is the cost and everything chained to it, each as a copy with no
// Also, in order. A nil or empty cost has none.
func (c *TapPermanentsCost) links() []*TapPermanentsCost {
	if c.Empty() {
		return nil
	}
	head := *c
	head.Also = nil
	out := []*TapPermanentsCost{&head}
	for _, r := range c.Also {
		if l := tapCostFromRef(r); l != nil {
			out = append(out, l)
		}
	}
	return out
}

// WithAlso is the cost with `other` chained after it, as a new value (the
// catalog's costs are shared and never edited). Either may be empty. A
// second convoke, or a second improvise, adds nothing (CR 702.51b), so it
// is dropped; two waterbends are two costs.
func (c *TapPermanentsCost) WithAlso(other *TapPermanentsCost) *TapPermanentsCost {
	if c.Empty() {
		return other
	}
	if other.Empty() {
		return c
	}
	ls := c.links()
	cp := *c
	for _, o := range other.links() {
		dup := false
		for _, l := range ls {
			if l.Extra == "" && o.Extra == "" && l.Key == o.Key {
				dup = true
				break
			}
		}
		if !dup {
			ls = append(ls, o)
			cp.Also = append(append([]TapCostRef(nil), cp.Also...), TapCostRef{Key: o.Key, Extra: o.Extra, InCost: o.InCost})
		}
	}
	return &cp
}

// ChainLabel is the clause as the picker shows it: each cost's label,
// joined, so a Hama cast of a convoke spell reads "Waterbend {3} +
// Convoke".
func (c *TapPermanentsCost) ChainLabel() string {
	if c.Empty() {
		return ""
	}
	ls := c.links()
	parts := make([]string, len(ls))
	for i, l := range ls {
		parts[i] = l.Label
	}
	return strings.Join(parts, " + ")
}

// ChainSpecs is every link's spec, for the view's candidate walk.
func (c *TapPermanentsCost) ChainSpecs() []*TargetSpec {
	var out []*TargetSpec
	for _, l := range c.links() {
		out = append(out, l.Spec)
	}
	return out
}

// tapCostSize is the number of mana symbols a parsed cost still owes.
func tapCostSize(p ParsedCost) int { return p.Generic + len(p.Required) }

// tapChainBudget is the most permanents a chained cast may tap: each
// link's own ceiling, measured against `total` (the cost after the
// modifiers and before any tap). The split decides whether the actual
// list fits; this only bounds its length.
func tapChainBudget(c *TapPermanentsCost, total ParsedCost, x int) int {
	n := 0
	for _, l := range c.links() {
		if l.InCost {
			n += WaterbendBudget(l, total, x)
		} else {
			n += tapPermanentsBudget(l, total, x)
		}
	}
	return n
}

// splitTapPayersLocked shares the tapped permanents between the chained
// costs. groups[i] pays links()[i]; ok is false when some permanent can
// pay none of them, or pays for nothing under the one it was given (a tap
// for a symbol that is not there).
//
// Costs are served in order. Each takes the permanents only it can use
// first, then the ones the later costs want least, and only those that
// still reduce what it owes; a creature with fewer colours goes first
// when a later cost is convoke, whose colour clause may want the others.
//
// `cost` is what the cast owes before any tap (the modifiers applied).
// Caller must hold g.mu.
func (g *Game) splitTapPayersLocked(c *TapPermanentsCost, cost ParsedCost, payers []Card, x int) ([][]Card, bool) {
	ls := c.links()
	groups := make([][]Card, len(ls))
	if len(ls) == 0 {
		return groups, len(payers) == 0
	}
	match := make([][]bool, len(payers))
	for i, p := range payers {
		match[i] = make([]bool, len(ls))
		for k, l := range ls {
			match[i][k] = g.specMatchLocked(SourceChooser(p.Controller), l.Spec,
				TargetRef{Kind: TargetCard, ID: p.InstanceID}, false)
		}
	}
	taken := make([]bool, len(payers))
	cur := cost
	for k, l := range ls {
		var cands []int
		for i := range payers {
			if !taken[i] && match[i][k] {
				cands = append(cands, i)
			}
		}
		laterColor := false
		for _, lt := range ls[k+1:] {
			laterColor = laterColor || lt.ColorClause
		}
		later := func(i int) int {
			n := 0
			for kk := k + 1; kk < len(ls); kk++ {
				if match[i][kk] {
					n++
				}
			}
			return n
		}
		sort.SliceStable(cands, func(a, b int) bool {
			la, lb := later(cands[a]), later(cands[b])
			if la != lb {
				return la < lb
			}
			if laterColor {
				return len(payers[cands[a]].EffectiveColors()) < len(payers[cands[b]].EffectiveColors())
			}
			return false
		})
		idle := tapCostSize(tapPermanentsAdjusted(cur, l, nil, x))
		var group []Card
		var chosen []int
		for _, i := range cands {
			try := append(append([]Card(nil), group...), payers[i])
			if idle-tapCostSize(tapPermanentsAdjusted(cur, l, try, x)) > len(group) {
				group = try
				chosen = append(chosen, i)
			}
		}
		for _, i := range chosen {
			taken[i] = true
		}
		groups[k] = group
		cur = tapPermanentsAdjusted(cur, l, group, x)
	}
	for i := range payers {
		if !taken[i] {
			return groups, false
		}
	}
	return groups, true
}

// tapPermanentsAdjustedLocked is tapPermanentsAdjusted for a cost that
// may be chained: each link is applied with the share the split gave it.
// Caller must hold g.mu.
func (g *Game) tapPermanentsAdjustedLocked(cost ParsedCost, c *TapPermanentsCost, payers []Card, x int) ParsedCost {
	if c.Empty() || !c.HasChain() {
		return tapPermanentsAdjusted(cost, c, payers, x)
	}
	groups, _ := g.splitTapPayersLocked(c, cost, payers, x)
	for k, l := range c.links() {
		cost = tapPermanentsAdjusted(cost, l, groups[k], x)
	}
	return cost
}

// tapChainPaysLocked reports whether the announced taps fit the chained
// cost: every tapped permanent pays for something under one of the
// costs. `total` is the cost the cast owes before the taps. Caller must
// hold g.mu.
func (g *Game) tapChainPaysLocked(c *TapPermanentsCost, total ParsedCost, ids []uuid.UUID, x int) bool {
	_, ok := g.splitTapPayersLocked(c, total, g.tapPermanentsPayersLocked(ids), x)
	return ok
}

// tapConvokersForCastLocked is who tapped for convoke on this cast, for
// "each creature that convoked this spell" (#3172). Caller must hold g.mu.
func (g *Game) tapConvokersForCastLocked(p *Player, card Card, tapCost *TapPermanentsCost, params CastSpellParams) []uuid.UUID {
	if tapCost.Empty() || len(params.TapIDs) == 0 {
		return nil
	}
	if !tapCost.HasChain() {
		if tapCost.Key == "convoke" {
			return append([]uuid.UUID(nil), params.TapIDs...)
		}
		return nil
	}
	base, _, err := g.printedCostLocked(p, card, params)
	if err != nil {
		return nil
	}
	untapped := params
	untapped.TapIDs, untapped.DelveIDs = nil, nil
	total, err := g.costAfterModifiersLocked(base, p, card, untapped)
	if err != nil {
		return nil
	}
	groups, _ := g.splitTapPayersLocked(tapCost, total, g.tapPermanentsPayersLocked(params.TapIDs), params.XValue)
	var out []uuid.UUID
	for k, l := range tapCost.links() {
		if l.Key == "convoke" {
			for _, c := range groups[k] {
				out = append(out, c.InstanceID)
			}
		}
	}
	return out
}
