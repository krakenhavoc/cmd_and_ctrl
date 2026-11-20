package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorilla Berserkers — Creature — Ape Berserker {3}{G}{G}, 2/3 (Alliances, #5769):
//
//	"Trample; rampage 2 (Whenever this creature becomes blocked, it gets
//	 +2/+2 until end of turn for each creature blocking it beyond the
//	 first.)
//	 This creature can't be blocked except by three or more
//	 creatures."
//
// Trample is the engine's keyword and rampage its keyword trigger (CR
// 702.23, game/rampage.go). The semicolon is old wording for a comma,
// which the importer's line scan reads as one since #5791; the keywords
// are declared here as well, and the merge keeps one instance. "Can't be blocked except by three or more creatures" is a block
// minimum (MinBlockers), so a declaration that blocks it names at least
// three creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6373c66d-edfc-44cb-af2a-2273e0ba53e0",
		Name:            "Gorilla Berserkers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "rampage 2"},
		BlockRules:      []game.BlockRule{effects.MinBlockers(effects.OnSelf(), 3)},
	})
}
