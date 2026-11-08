package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frilled Sparkshooter — Creature — Lizard Archer {3}{R}, 3/3
// (Bloomburrow, #4467):
//
//	"Reach, menace
//	 This creature enters with a +1/+1 counter on it if an opponent lost
//	 life this turn."
//
// Reach and menace are printed keywords. The counter is Cindering
// Cutthroat's CR 614.1c self-replacement, sharing its count
// (oneIfAnOpponentLostLifeThisTurn): one if any opponent of its
// controller lost life this turn, read as the Sparkshooter enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69038520-7156-48d9-aae8-ba3cc76a5993",
		Name:            "Frilled Sparkshooter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "menace"},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, oneIfAnOpponentLostLifeThisTurn,
				"Frilled Sparkshooter enters with a +1/+1 counter if an opponent lost life this turn"),
		},
	})
}
