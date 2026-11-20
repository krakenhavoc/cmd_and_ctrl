package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exultant Cultist — Creature — Human Wizard {2}{U}, 2/2 (Eldritch Moon,
// #5541):
//
//	"When this creature dies, draw a card."
//
// Spore Crawler's dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e802442e-6a78-4c15-9034-fc8872470208",
		Name:         "Exultant Cultist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Exultant Cultist — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
