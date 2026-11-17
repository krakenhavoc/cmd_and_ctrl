package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kingfisher — Creature — Bird {3}{U}, 2/2 (Urza's Destiny, #5621):
//
//	"Flying
//	 When this creature dies, draw a card."
//
// Ithilien Kingfisher's dies trigger on a 2/2. It goes on the stack
// once the Bird is in the graveyard, and the player who last controlled
// it draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2a70f3d7-9b5c-455e-8c4b-55870d77bd7d",
		Name:            "Kingfisher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Kingfisher — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
