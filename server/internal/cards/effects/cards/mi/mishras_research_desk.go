package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "95ac482c-c6ad-49b4-937e-460241ae355f",
		Name:         "Mishra's Research Desk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label: "{1}, {T}, Sacrifice this artifact: Exile the top two cards of your library. Choose one of them. Until the end of your next turn, you may play that card.",
				Cost:  effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
				// Exile-and-play card advantage answers nothing on the stack.
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.S82Bro13ExileChooseOnePlay(ctx, 2)
				})),
			},
			effects.Unearth("{1}{R}"),
		},
	})
}
