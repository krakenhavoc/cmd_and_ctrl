package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raugrin Triome — Land — Island Mountain Plains:
//
//	"({T}: Add {U}, {R}, or {W}.)
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
		OracleID:     "c7fa1dda-9312-4ec8-82cd-a1ba7bc33497",
		Name:         "Raugrin Triome",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|R|W}",
			Label:    "Add {U}, {R}, or {W}",
		}},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
