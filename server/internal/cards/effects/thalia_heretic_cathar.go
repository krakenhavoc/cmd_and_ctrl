package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thalia, Heretic Cathar — Legendary Creature — Human Soldier {2}{W},
// 3/2 (EDHREC rank 1473):
//
//	"First strike
//	 Creatures and nonbasic lands your opponents control enter
//	 tapped."
//
// The tempo hatebear. Kismet's replacement narrowed to creatures and
// NONBASIC lands: the entering permanent's controller must not be
// Thalia's, and "nonbasic" is the Basic supertype read off the
// effective characteristic — a land Urborg made a Swamp is still
// nonbasic, a Snow-Covered Plains is still basic. Every entry path
// runs the pipeline (cast, played, fetched, reanimated, a token), so
// an opponent's fetched shockland and their Zombie tokens all arrive
// tapped, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13be47e4-9da2-4f9e-baf0-db96f7777ccb",
		Name:            "Thalia, Heretic Cathar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements: []game.ReplacementEffect{
			opponentsCreaturesAndNonbasicLandsEnterTapped("Thalia, Heretic Cathar: enters tapped"),
		},
	})
}
