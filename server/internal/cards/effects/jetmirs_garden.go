package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jetmir's Garden — Land — Mountain Forest Plains:
//
//	"({T}: Add {R}, {G}, or {W}.)
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
		OracleID:     "f5896356-5744-4f7e-a4e5-1cc36dde5958",
		Name:         "Jetmir's Garden",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|G|W}",
			Label:    "Add {R}, {G}, or {W}",
		}},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
