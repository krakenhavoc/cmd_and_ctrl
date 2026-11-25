package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aeon Engine — Artifact {5} (The Dark, #6223):
//
//	"This artifact enters tapped.
//	 {T}, Exile this artifact: Reverse the game's turn order. (For
//	 example, if play had proceeded clockwise around the table, it now
//	 goes counterclockwise.)"
//
// The enters-tapped clause is the CR 614 self-replacement Moss Diamond
// uses. The ability flips Game.TurnOrderReversed (#6223): the next turn,
// every priority pass and every APNAP walk (CR 101.4) go the other way
// round the table from then on. A second Aeon Engine turns it back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af471455-6b9f-4c14-bb90-a74a44ecd247",
		Name:         "Aeon Engine",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Exile this artifact: Reverse the game's turn order.",
			// Reordering the turns answers nothing on the stack.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.ExileThis()),
			Effect: func(g *game.Game, _ *game.StackItem) error {
				g.ReverseTurnOrderForEffect()
				return nil
			},
		}},
	})
}
