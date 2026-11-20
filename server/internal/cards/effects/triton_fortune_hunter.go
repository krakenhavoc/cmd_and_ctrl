package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triton Fortune Hunter — Creature — Merfolk Soldier {2}{U}, 2/2 (Theros, #5561):
//
//	"Heroic — Whenever you cast a spell that targets this creature, draw a card."
//
// The shared heroic trigger (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f84637f0-1a04-4fd1-85d9-8e8f5d444800",
		Name:         "Triton Fortune Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			heroic("Triton Fortune Hunter — draw a card (heroic)", Do(DrawCards{N: 1})),
		},
	})
}
