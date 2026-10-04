package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savai Triome — Land — Mountain Plains Swamp:
//
//	"({T}: Add {R}, {W}, or {B}.)
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
		OracleID:     "00625242-9348-4ef4-b975-f2ac82fee21d",
		Name:         "Savai Triome",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|W|B}",
			Label:    "Add {R}, {W}, or {B}",
		}},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
