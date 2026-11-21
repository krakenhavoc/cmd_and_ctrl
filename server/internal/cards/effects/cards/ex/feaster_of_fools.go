package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feaster of Fools — Creature — Demon {4}{B}{B}, 3/3 (Modern Horizons, #5519):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you tap
//	 while casting this spell pays for {1} or one mana of that creature's
//	 color.)
//	 Flying
//	 Devour 2 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with twice that many +1/+1 counters on it.)"
//
// Convoke is the shared tap cost, devour Mycoloth's replacement at 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3e3d4f28-93a0-4313-a9a0-48fd06540cf8",
		Name:            "Feaster of Fools",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         effects.Convoke(),
		Replacements:    []game.ReplacementEffect{effects.Devour("Feaster of Fools", 2)},
	})
}
