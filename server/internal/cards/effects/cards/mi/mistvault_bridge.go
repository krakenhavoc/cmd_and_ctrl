package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistvault Bridge — Artifact Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {U} or {B}."
//
// Rustvale Bridge's shape with another colour pair.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "33ee23bc-6327-4a54-a704-dfd83be36bb5",
		Name:            "Mistvault Bridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities:   []effects.ManaAbility{effects.DualManaAbility("U", "B")},
	})
}
