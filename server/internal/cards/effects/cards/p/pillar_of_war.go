package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pillar of War — Artifact Creature — Golem {3}, 3/3 (Born of the Gods, S82 OLD
// b379, #5559):
//
//	"Defender
//	 As long as this creature is enchanted, it can attack as though it didn't
//	 have defender."
//
// Backstreet Bruiser's conditional waiver (CR 702.3b), the condition being an
// Aura attached to it (a bestowed card counts), read on every check.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "50b931f9-f843-4129-a4c8-2bb7a051c063",
		Name:            "Pillar of War",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{effects.AvB18CanAttackAsThoughNoDefenderWhile(
			"As long as this creature is enchanted, it can attack as though it didn't have defender.",
			effects.S82OldB379IsEnchanted)},
	})
}
