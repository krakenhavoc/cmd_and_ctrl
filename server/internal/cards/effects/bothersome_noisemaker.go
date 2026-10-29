package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bothersome Noisemaker — Creature — Goblin Bard {1}{R}, 2/2 (The Hobbit):
//
//	"Whenever you cast a noncreature spell, amass Goblins 1."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "978e07fd-8b8c-4eab-850e-e4c0e712251e",
		Name:         "Bothersome Noisemaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Bothersome Noisemaker — amass Goblins 1", Do(Amass{Subtype: "Goblin", N: 1})),
		},
	})
}
