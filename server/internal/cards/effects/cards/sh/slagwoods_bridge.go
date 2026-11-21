package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slagwoods Bridge — Artifact Land:
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {R} or {G}."
//
// Darksteel Citadel's indestructible artifact land with Canyon Slough's
// two-colour mana ability and enters-tapped replacement. The indestructible
// keyword is printed, so the destroy verbs and lethal damage honour it
// (CR 702.12b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e040a8e6-b90c-42d1-a1b1-771d954c61ab",
		Name:            "Slagwoods Bridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities:   []effects.ManaAbility{effects.DualManaAbility("R", "G")},
	})
}
