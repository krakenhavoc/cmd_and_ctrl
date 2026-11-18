package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drossforge Bridge — Artifact Land:
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {B} or {R}."
//
// Darkmoss Bridge's shape in black and red.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44b83535-fdbf-4307-bf53-ca20470a768d",
		Name:            "Drossforge Bridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:   []ManaAbility{dualManaAbility("B", "R")},
	})
}
