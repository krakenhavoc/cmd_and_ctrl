package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tanglepool Bridge — Artifact Land:
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {G} or {U}."
//
// Darksteel Citadel's indestructible artifact land with Canyon Slough's
// two-colour mana ability and enters-tapped replacement. The indestructible
// keyword is printed, so the destroy verbs and lethal damage honour it
// (CR 702.12b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "29cd8a7c-108a-43d1-af63-f603a27c24f2",
		Name:            "Tanglepool Bridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:   []ManaAbility{dualManaAbility("G", "U")},
	})
}
