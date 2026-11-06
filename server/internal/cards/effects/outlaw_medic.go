package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outlaw Medic — Creature — Human Rogue {1}{W}, 1/3 (Outlaws of Thunder
// Junction, #4468):
//
//	"Lifelink
//	 When this creature dies, draw a card."
//
// Lifelink is the engine's keyword (CR 702.15b); the dies trigger
// (CR 603.10a) draws for its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "637be5f7-ab39-4e3a-a5eb-c6e48fca00a9",
		Name:            "Outlaw Medic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Outlaw Medic — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
