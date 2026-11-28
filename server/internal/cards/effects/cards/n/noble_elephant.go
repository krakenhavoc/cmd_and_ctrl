package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noble Elephant — Creature — Elephant {3}{W}, 2/2 (Mirage, S82 OLD b807):
//
//	"Trample; banding (Any creatures with banding, and up to one without, can
//	 attack in a band. Bands are blocked as a group. If any creatures with
//	 banding you control are blocking or being blocked by a creature, you
//	 divide that creature's combat damage, not its controller, among any of
//	 the creatures it's being blocked by or is blocking.)"
//
// Trample and banding are both the engine's (game/banding.go, ADR 0224), as on
// Mesa Pegasus.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99dd6f15-855f-431f-a3c8-31f1421af88f",
		Name:            "Noble Elephant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordBanding},
	})
}
