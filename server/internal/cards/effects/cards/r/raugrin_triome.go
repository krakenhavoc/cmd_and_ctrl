package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c7fa1dda-9312-4ec8-82cd-a1ba7bc33497",
		Name:         "Raugrin Triome",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|R|W}",
			Label:    "Add {U}, {R}, or {W}",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
