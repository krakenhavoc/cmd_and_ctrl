package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Planar Ally — Creature — Angel {3}{W}{W}, 3/3:
//
//	"Flying
//	 Whenever this creature attacks, venture into the dungeon."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ae30bd5-b638-4647-b59d-ffe08e77c182",
		Name:            "Planar Ally",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(ventureLabel("Planar Ally"), Venture),
		},
	})
}
