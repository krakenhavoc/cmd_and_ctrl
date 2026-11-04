package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prescient Chimera — Creature — Chimera {3}{U}{U}, 3/4 (Jumpstart
// Azorius deck request, #4894):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, scry 1."
//
// A cast trigger (Jace's Sanctum's): it goes on the stack above the
// spell and resolves first, whether or not the spell itself later
// resolves. Copies are not cast and do not trigger it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "af199b80-d5d8-417a-97d5-459a5bd48b22",
		Name:            "Prescient Chimera",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Prescient Chimera — scry 1", Do(Scry{N: 1})),
		},
	})
}
