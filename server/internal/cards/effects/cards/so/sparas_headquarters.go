package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spara's Headquarters — Land — Forest Plains Island:
//
//	"({T}: Add {G}, {W}, or {U}.)
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
		OracleID:     "3123ec89-8e95-4761-ba17-747ec667509f",
		Name:         "Spara's Headquarters",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|W|U}",
			Label:    "Add {G}, {W}, or {U}",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
