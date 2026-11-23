package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raffine's Tower — Land — Plains Island Swamp:
//
//	"({T}: Add {W}, {U}, or {B}.)"
//	"This land enters tapped."
//	"Cycling {3}"
//
// The Esper tri-land. Enters-tapped is a real CR 614
// self-replacement and the three-colour pipe gives one picker rather
// than three menu entries.
//
// Cycling arrived with #660. It is an activated ability that
// functions from hand (CR 702.29a) — "{3}, Discard this card: Draw a
// card" — not an alternative way to cast the land, which is what the
// caveat this file used to carry said it was.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e9ef5ef-6aed-4d3e-a59b-9e3dc8740b1b",
		Name:         "Raffine's Tower",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B}",
			Label:    "Add {W}, {U}, or {B}",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
