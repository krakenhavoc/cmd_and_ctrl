package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bombur, Gentle Dreamer — Legendary Creature — Dwarf Bard {2}{R}, 5/3:
//
//	"Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 Bombur doesn't untap during your untap step unless you have an
//	 enduring story."
//
// Storied is the engine's (CR 702.195a). The untap clause is a CR 502.3
// restriction asked at each of your untap steps, so the first untap step
// after the story is told untaps him, and every one after.
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:              "925ce68e-9323-4985-8fef-0d3f36e0d74c",
		Name:                  "Bombur, Gentle Dreamer",
		Completeness:          CompletenessFull,
		PrintedKeywords:       []string{game.KeywordStoried},
		UntapStepRestrictions: []game.UntapStepRestriction{DoesntUntapUnlessEnduringStory()},
	})
}
