package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mounted Dreadknight — Creature — Vampire Knight {4}{R}, 5/4 (Innistrad:
// Midnight Hunt, #5497):
//
//	"Trample
//	 This creature enters with a +1/+1 counter on it if an opponent lost
//	 life this turn."
//
// Frilled Sparkshooter's CR 614.1c self-replacement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5e42be3-6bfc-4905-9237-cb8d53e2349e",
		Name:            "Mounted Dreadknight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, oneIfAnOpponentLostLifeThisTurn,
				"Mounted Dreadknight enters with a +1/+1 counter if an opponent lost life this turn"),
		},
	})
}
