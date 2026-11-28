package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teremko Griffin — Creature — Griffin {3}{W}, 2/2 (Mirage, S82 OLD b807):
//
//	"Flying; banding (Any creatures with banding, and up to one without, can
//	 attack in a band. Bands are blocked as a group. If any creatures with
//	 banding you control are blocking or being blocked by a creature, you
//	 divide that creature's combat damage, not its controller, among any of
//	 the creatures it's being blocked by or is blocking.)"
//
// Flying and banding are both the engine's (game/banding.go, ADR 0224), as on
// Mesa Pegasus.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a00cb1fa-a196-4403-9008-532dbd15769e",
		Name:            "Teremko Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordBanding},
	})
}
