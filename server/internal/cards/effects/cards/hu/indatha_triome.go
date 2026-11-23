package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indatha Triome — Land — Plains Swamp Forest:
//
//	"({T}: Add {W}, {B}, or {G}.)"
//	"This land enters tapped."
//	"Cycling {3}"
//
// The Ikoria Triome for the Abzan wedge. Its three real basic land
// types — same reasoning as Ketria Triome — are read straight off the
// printed TypeLine by IsLandWithSubtype and need nothing here.
//
// The mana ability is declared rather than derived for the same
// reason every nonbasic dual in the catalog declares it: the engine's
// synthetic land ability only fires for lands with the BASIC
// supertype, so a Triome's reminder-text ability has to be spelled
// out or the land taps for nothing.
//
// Cycling {3} is the same activated ability every Triome carries
// (CR 702.29a) — an ordinary hand ability, not an alternative cast
// cost (ADR 0062).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec2b3779-55f7-4169-aa66-6312fb52721f",
		Name:         "Indatha Triome",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|B|G}",
			Label:    "Add {W}, {B}, or {G}",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
