package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Dustwasp — Creature — Insect {3}{G}{G}, 3/3 (Planar Chaos, #5595):
//
//	"Flying
//	 Suspend 4—{1}{G} (Rather than cast this card from your hand, you may pay
//	 {1}{G} and exile it with four time counters on it. At the beginning of
//	 your upkeep, remove a time counter. When the last is removed, you may
//	 cast it without paying its mana cost. It has haste.)"
//
// Errant Ephemeron's shape: flying is printed data, and suspend is the
// Suspend special action (CR 702.62a), which also gives the creature cast
// from it haste.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5f4e45c1-2bd2-4741-bb97-b6a324c574ba",
		Name:            "Giant Dustwasp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(4, "{1}{G}"),
		},
	})
}
