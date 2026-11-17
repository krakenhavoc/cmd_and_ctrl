package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blastoderm — Creature — Beast {2}{G}{G}, 5/5:
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Fading 3 (This creature enters with three fade counters on it. At
//	 the beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)"
//
// Shroud is the engine's keyword; fading is Fading/FadingUpkeep
// (fading.go, CR 702.32a), so it attacks for three turns and is
// sacrificed at the fourth upkeep.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6ee8bd6-0207-46ee-9892-c1d1b23b934e",
		Name:            "Blastoderm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Replacements:    []game.ReplacementEffect{Fading(3)},
		Triggered:       []game.TriggeredAbility{FadingUpkeep()},
	})
}
