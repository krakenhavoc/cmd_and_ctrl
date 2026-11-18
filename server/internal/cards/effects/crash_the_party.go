package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crash the Party — Instant {5}{G} (New Capenna Commander, #5491):
//
//	"Create a tapped 4/4 green Rhino Warrior creature token for each tapped
//	 creature you control."
//
// The count is taken as the spell resolves, over the creatures you control
// that are tapped then (CR 608.2h), and the tokens enter tapped, so they do
// not raise the count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "337e322b-6144-4e0b-b998-df475166fcf6",
		Name:         "Crash the Party",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsCreature() && c.Tapped {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return CreateTokenAdvanced{
				Controller: uuid.Nil,
				Spec:       Token(TokenCard("4/4 green Rhino Warrior")).EntersTapped(),
				N:          n,
			}.Apply(ctx)
		},
	})
}
