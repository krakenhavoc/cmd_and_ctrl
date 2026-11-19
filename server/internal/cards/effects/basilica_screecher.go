package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basilica Screecher — Creature — Bat {1}{B}, 1/2 (Ravnica: Clue Edition):
//
//	"Flying
//	 Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each
//	 opponent loses 1 life and you gain that much life.)"
//
// Flying is the engine keyword; extort is the shared Extort trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "752ebcba-97fb-43a3-8811-d244164b3e16",
		Name:            "Basilica Screecher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{Extort("Basilica Screecher")},
	})
}
