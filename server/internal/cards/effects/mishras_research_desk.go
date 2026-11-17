package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mishra's Research Desk — Artifact {1} (The Brothers' War, #5482):
//
//	"{1}, {T}, Sacrifice this artifact: Exile the top two cards of your
//	 library. Choose one of them. Until the end of your next turn, you may
//	 play that card.
//	 Unearth {1}{R} (…)"
//
// Heroes' Hangout's Date Night body (s82Bro13ExileChooseOnePlay): both cards
// are exiled face up, the controller chooses one, and only it may be played
// until the end of their next turn. Unearth is the shared keyword; a Desk
// that comes back is exiled at the next end step or if it would leave.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95ac482c-c6ad-49b4-937e-460241ae355f",
		Name:         "Mishra's Research Desk",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{1}, {T}, Sacrifice this artifact: Exile the top two cards of your library. Choose one of them. Until the end of your next turn, you may play that card.",
				Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
				// Exile-and-play card advantage answers nothing on the stack.
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: Do(Instruction(func(ctx *Context) error {
					return s82Bro13ExileChooseOnePlay(ctx, 2)
				})),
			},
			Unearth("{1}{R}"),
		},
	})
}
