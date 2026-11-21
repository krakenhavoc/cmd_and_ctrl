package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudskate — Creature — Illusion {1}{U}, 2/2:
//
//	"Flying
//	 Fading 3 (This creature enters with three fade counters on it. At
//	 the beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)"
//
// Flying is the engine's keyword; fading is Fading/FadingUpkeep
// (fading.go, CR 702.32a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "648e4072-b991-46fd-b057-1e807091dcdf",
		Name:            "Cloudskate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.Fading(3)},
		Triggered:       []game.TriggeredAbility{effects.FadingUpkeep()},
	})
}
