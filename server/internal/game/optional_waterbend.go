package game

import "github.com/google/uuid"

// optional_waterbend.go — an optional "you may waterbend {N}" additional
// cost (CR 701.67, #3999, ADR 0073's 2026-10-19 amendment): Secret of
// Bloodbending's "As an additional cost to cast this spell, you may
// waterbend {10}."
//
// It is an optional cost (ADR 0073) whose mana half is the {N}: chosen,
// it joins the total at CR 601.2f like a kicker, so the cost modifiers,
// the auto-tapper and every affordability check already price it. The
// tapping is the other half: while paying a waterbend cost the caster
// may tap untapped artifacts and creatures they control, each paying
// {1} of it (CR 701.67a). That is the one tap cost a cast carries
// (TapIDs), so a chosen optional waterbend IS the cast's tap cost,
// marked InCost as Hama's waterbend alternative cost is (#3989): its {N}
// is already in the total, and the taps pay only its generic, never the
// printed cost beside it (CR 701.67b). Unchosen, the cast has no tap
// cost and a tap_ids is refused like any stray one.

// optionalWaterbendLocked is the tap cost an announcement's chosen
// optional costs carry: the first chosen one with a Waterbend clause, as
// an InCost copy, or nil. Caller must hold g.mu.
func (g *Game) optionalWaterbendLocked(caster uuid.UUID, card Card, chosen []int) *TapPermanentsCost {
	if len(chosen) == 0 {
		return nil
	}
	costs := g.OptionalCostsForCastLocked(caster, card)
	for _, i := range chosen {
		if i < 0 || i >= len(costs) || costs[i].Waterbend.Empty() {
			continue
		}
		tc := *costs[i].Waterbend
		tc.InCost = true
		return &tc
	}
	return nil
}
