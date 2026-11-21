package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charcoal Diamond — Artifact {2} (EDHREC rank 694):
//
//	"This artifact enters tapped.
//	 {T}: Add {B}."
//
// The Mirage diamond: a two-mana rock that costs you the turn it
// lands. Worn Powerstone's shape exactly — a real CR 614
// self-replacement for the tapped entry (no tap event, no window in
// which it was untapped) and one fixed-colour tap ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1386d111-a2a7-4df1-91d7-947664126989",
		Name:         "Charcoal Diamond",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
