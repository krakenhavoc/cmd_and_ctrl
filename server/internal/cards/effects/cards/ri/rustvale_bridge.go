package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rustvale Bridge — Artifact Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {R} or {W}."
//
// Thornglint Bridge's shape with another colour pair.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a3faf70d-c034-4692-9e92-1922029e3852",
		Name:            "Rustvale Bridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities:   []effects.ManaAbility{effects.DualManaAbility("R", "W")},
	})
}
