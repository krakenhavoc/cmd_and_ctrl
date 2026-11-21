package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b5e42be3-6bfc-4905-9237-cb8d53e2349e",
		Name:            "Mounted Dreadknight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, effects.OneIfAnOpponentLostLifeThisTurn,
				"Mounted Dreadknight enters with a +1/+1 counter if an opponent lost life this turn"),
		},
	})
}
