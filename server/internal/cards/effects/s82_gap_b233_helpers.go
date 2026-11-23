package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b233_helpers.go — GAP batch 233 (#5948): Stick Together's party
// choice and Sculpted Sunburst's power-bounded choice.

// GapB233PartyTypes are the four creature types of a party (CR 700.8).
var GapB233PartyTypes = [4]string{"Cleric", "Rogue", "Warrior", "Wizard"}

// GapB233IsParty reports whether the picked creatures are a party: each one
// stands for a party type of its own, so they can be matched one to one onto
// distinct types they have (a changeling has every type, CR 702.73a). A set
// of up to four that can be matched is a legal "up to one each of Cleric,
// Rogue, Warrior, and Wizard" choice.
func GapB233IsParty(picked []game.Card) bool {
	if len(picked) == 0 || len(picked) > len(GapB233PartyTypes) {
		return false
	}
	var owner [4]int
	for i := range owner {
		owner[i] = -1
	}
	var try func(c int, seen *[4]bool) bool
	try = func(c int, seen *[4]bool) bool {
		for t, name := range GapB233PartyTypes {
			if seen[t] || !picked[c].HasSubtype(name) {
				continue
			}
			seen[t] = true
			if owner[t] < 0 || try(owner[t], seen) {
				owner[t] = c
				return true
			}
		}
		return false
	}
	for c := range picked {
		var seen [4]bool
		if !try(c, &seen) {
			return false
		}
	}
	return true
}

// GapB233PartyCandidates are the creatures `seat` controls that have a party
// type, through the layered view.
func GapB233PartyCandidates(g *game.Game, seat uuid.UUID) []uuid.UUID {
	g.RecomputeLayersIfStaleLocked()
	var out []uuid.UUID
	for _, id := range PermanentsControlledByMatching(g, seat, Creature()) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, name := range GapB233PartyTypes {
			if c.HasSubtype(name) {
				out = append(out, id)
				break
			}
		}
	}
	return out
}

// GapB233PartyLeg is the registered set rule for one seat's party pick.
var GapB233PartyLeg = game.RegisterOrderedPickLegRule("choose-cards/gap-b233-stick-together-party", GapB233IsParty)
