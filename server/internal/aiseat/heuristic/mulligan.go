package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// mulligan.go is #2693: an opening hand at the land floor is kept only
// if it can cast something soon. Review game 2 (06e98afa) kept Mountain,
// Exotic Orchard and five spells of three to seven mana on "2 lands in
// 7", and missed its third and fourth land drops with nothing to do.

// castableSoon reports whether the hand holds a nonland spell whose mana
// value is at most lands + KeepCastReach and whose coloured pips the
// hand's lands can make.
//
// Colours are read off each land's mana abilities (producedColors). A
// land whose abilities name no colour it can be read for, and is not a
// colourless-only land, counts as any colour: a fetch land (no mana
// ability on the view), Exotic Orchard ("any color an opponent's land
// could produce"). That errs toward keeping, which is the old
// behaviour. The reach is generic mana only: the land the hand is
// waiting to draw has no known colour.
//
// A card with no mana cost cannot be cast for one and does not count;
// an X spell counts at X = 0, as manaShort reads it.
func (p *Policy) castableSoon(hand []protocol.CardView, lands int) bool {
	have := map[string]int{}
	wild := 0
	for i := range hand {
		c := &hand[i]
		if !isLand(c) {
			continue
		}
		colors := producedColors(c)
		switch {
		case len(colors) > 0:
			for k := range colors {
				have[k]++
			}
		case !colorlessOnly(c):
			wild++
		}
	}
	for i := range hand {
		c := &hand[i]
		if isLand(c) || c.ManaCost == "" {
			continue
		}
		if manaValue(c.ManaCost, 0) > lands+p.cfg.KeepCastReach {
			continue
		}
		gap := 0
		for k, n := range colorPips(c.ManaCost) {
			if d := n - have[k]; d > 0 {
				gap += d
			}
		}
		if gap <= wild {
			return true
		}
	}
	return false
}

// colorlessOnly reports whether every repeatable mana ability a land has
// makes only {C}, like Field of the Dead's. Such a
// land pays generic costs and no coloured pip.
func colorlessOnly(c *protocol.CardView) bool {
	seen := false
	for i := range c.ManaAbilities {
		ab := &c.ManaAbilities[i]
		if !ab.TapCost || ab.SacrificeCost || ab.ExileSelf || ab.AddsNoMana {
			continue
		}
		if ab.Produced == "" {
			return false
		}
		for _, r := range ab.Produced {
			switch r {
			case 'W', 'U', 'B', 'R', 'G':
				return false
			}
		}
		seen = true
	}
	return seen
}
