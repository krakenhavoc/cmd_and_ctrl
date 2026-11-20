package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Professor of Zoomancy — Creature — Bear Druid {3}{G}, 4/3 (Strixhaven:
// School of Mages, #5502):
//
//	"When this creature enters, create a 1/1 black and green Pest creature
//	 token with "When this token dies, you gain 1 life.""
//
// The Strixhaven Pest (PestToken, with its own dies trigger), made by an
// enters trigger on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a22b8220-b41d-4b58-aa68-2c65ce37c517",
		Name:         "Professor of Zoomancy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Professor of Zoomancy — create a 1/1 Pest that gains you 1 life when it dies",
				Do(CreateToken{Template: PestToken(), N: 1})),
		},
	})
}
