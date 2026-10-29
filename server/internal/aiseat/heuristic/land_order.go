package heuristic

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func (p *Policy) landDropValue(st *state, m legal.Move) (float64, string) {
	v := p.cfg.LandValue
	if !p.cfg.OrderLandDrops {
		return v, "land drop"
	}
	cp := decode[castParams](m.Params)
	c := st.castSource(cp.InstanceID)
	if c == nil {
		return v, "land drop"
	}
	if c.Purpose != nil && c.Purpose.ReturnsLandOnEntry {
		other := false
		for _, land := range st.bf {
			if land.Controller == st.me && isLand(land) {
				other = true
				break
			}
		}
		if !other {
			// Still play it if it is the only land; this orders drops,
			// rather than changing the keep/mulligan or landfall policy.
			return v - 1, "land drop: bounce land has no other land to return"
		}
	}
	if landEntersUntapped(c) && p.landUnlocksSpell(st, c) {
		for _, other := range st.landChoices {
			if other.Kind != legal.KindLand || other.Source == m.Source {
				continue
			}
			alt := st.castSource(decode[castParams](other.Params).InstanceID)
			if alt != nil && alt.Purpose != nil && alt.Purpose.LandEntry == "tapped" {
				return v + 0.25, "land drop: unlocks a spell's mana cost"
			}
		}
	}
	return v, "land drop"
}

func landEntersUntapped(c *protocol.CardView) bool {
	if c.Purpose != nil && c.Purpose.LandEntry != "" {
		return c.Purpose.LandEntry == "untapped"
	}
	// Basic lands need no catalog registration. Other unknown and
	// conditional entries keep the old price; never parse oracle text.
	return typeLineWord(c.TypeLine, "basic") && isLand(c)
}

// landUnlocksSpell reuses the turn planner's coloured, restricted and
// filter mana model. It checks ordinary printed costs in our hand,
// excluding additional costs and X. This is a sequencing estimate,
// not a future legal-move enumeration: taxes, discounts, targeting and
// other entry replacements can still change what is offered afterwards.
func (p *Policy) landUnlocksSpell(st *state, land *protocol.CardView) (unlocks bool) {
	if v, ok := st.landUnlocks[land.InstanceID]; ok {
		return v
	}
	if st.landUnlocks == nil {
		st.landUnlocks = map[string]bool{}
	}
	defer func() { st.landUnlocks[land.InstanceID] = unlocks }()
	if st.seat == nil {
		return false
	}
	mm := p.manaModel(st)
	before := st.manaAvailable(mm)
	view := *st.view
	view.Battlefield.Cards = append(append([]protocol.CardView(nil), st.view.Battlefield.Cards...), *land)
	added := &view.Battlefield.Cards[len(view.Battlefield.Cards)-1]
	added.Controller, added.Tapped = st.me, false
	added.SummoningSick = isCreature(added) && !hasKeyword(added, "haste")
	future := *st
	future.view = &view
	after := future.manaAvailable(mm)
	for i := range st.seat.Hand.Cards {
		card := &st.seat.Hand.Cards[i]
		if isLand(card) || card.ManaCost == "" || card.AdditionalCost != nil || strings.Contains(card.ManaCost, "{X}") {
			continue
		}
		cost := parseManaCost(card.ManaCost, 0)
		cost.admits = st.castAdmits(card, "hand")
		if _, ok := payMana(append([]manaUnit(nil), before...), cost, nil); ok {
			continue
		}
		if _, ok := payMana(append([]manaUnit(nil), after...), cost, nil); ok {
			return true
		}
	}
	return false
}
