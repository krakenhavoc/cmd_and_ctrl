package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caprichrome — Artifact Creature — Goat {3}{W}, 2/2 (Modern Horizons 2,
// #5501):
//
//	"Flash
//	 Vigilance
//	 Devour artifact 1 (As this creature enters, you may sacrifice any number
//	 of artifacts. It enters with that many +1/+1 counters on it.)"
//
// Devour over artifacts (CR 702.82a): the sacrifice and the counters are
// one entry, and the entering Caprichrome is never a candidate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27831489-b444-4b7b-91db-feb54281bca9",
		Name:            "Caprichrome",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "vigilance"},
		Replacements: []game.ReplacementEffect{
			effects.DevourOver("Caprichrome", 1, "artifacts", func(c game.Card) bool { return c.IsArtifact() }),
		},
	})
}
