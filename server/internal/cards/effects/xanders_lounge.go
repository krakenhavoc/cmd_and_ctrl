package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Xander's Lounge — Land — Island Swamp Mountain:
//
//	"({T}: Add {U}, {B}, or {R}.)
//	 This land enters tapped.
//	 Cycling {3}"
//
// A Streets of New Capenna Triome — Indatha Triome's template with the
// colours swapped for Grixis. Its three basic land types are read
// straight off the printed TypeLine, so nothing here declares them;
// the mana ability is spelled out because the engine's synthetic land
// ability only fires for a card with the BASIC supertype, and Cycling
// {3} is the same ordinary hand ability every Triome carries
// (CR 702.29a, ADR 0062).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8291543f-d086-48aa-b2b7-5481ca8c9198",
		Name:         "Xander's Lounge",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|B|R}",
			Label:    "Add {U}, {B}, or {R}",
		}},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
