package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spore Crawler — Creature — Fungus {2}{G}, 3/2 (Innistrad: Crimson Vow):
//
//	"When this creature dies, draw a card."
//
// Buzz Bots' dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa7c13c0-350b-4ed6-8c16-590d2b988184",
		Name:         "Spore Crawler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Spore Crawler — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
