package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ziatora's Proving Ground — Land — Swamp Mountain Forest:
//
//	"({T}: Add {B}, {R}, or {G}.)
//	 This land enters tapped.
//	 Cycling {3}"
//
// A Triome — Xander's Lounge's template with this card's three
// colours. Its three basic land types are read straight off the
// printed TypeLine, so nothing here declares them; the mana ability is
// spelled out because the engine's synthetic land ability only fires
// for a card with the BASIC supertype, and Cycling {3} is the same
// ordinary hand ability every Triome carries (CR 702.29a, ADR 0062).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7e7b78c-c769-4720-8585-1874773eb342",
		Name:         "Ziatora's Proving Ground",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|R|G}",
			Label:    "Add {B}, {R}, or {G}",
		}},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
