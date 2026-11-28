package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Isle — Legendary Land (Mirage, S82 OLD batch 812, #5633):
//
//	"Phasing (This phases in or out before you untap during each of your untap
//	 steps. While it's phased out, it's treated as though it doesn't exist.)
//	 Teferi's Isle enters tapped.
//	 {T}: Add {U}{U}."
//
// Phasing is the engine's own keyword; the entry replacement is Mortuary Mire's
// (SelfEntersTapped), and the mana ability adds both mana at once (CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce55657d-d82f-4528-a83e-5cad7de111fd",
		Name:            "Teferi's Isle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPhasing},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}{U}",
			Label:    "Add {U}{U}",
		}},
	})
}
