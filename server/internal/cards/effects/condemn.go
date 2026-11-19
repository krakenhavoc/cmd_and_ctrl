package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Condemn — Instant {W} (Special Guests):
//
//	"Put target attacking creature on the bottom of its owner's library.
//	 Its controller gains life equal to its toughness."
//
// The creature's controller and toughness are read before it moves
// (CR 608.2h), and the life is gained once the move has settled: a
// commander its owner sends to the command zone (CR 903.9) still gives
// its controller the life, as the second sentence is not an "if you do".
func init() {
	Register(Spec{
		OracleID:     "377d31b5-0a48-4d5b-91bd-b5a62d4f7899",
		Name:         "Condemn",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 || targets[0].Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(targets[0].ID)
			if !ok {
				return nil
			}
			return PutIntoLibrary{
				Card:     targets[0].ID,
				ToBottom: true,
				ThenTail: withN(withCarry(ctx.Tail(condemnGainLife), c.Controller), c.CurrentToughness()),
			}.Apply(ctx)
		},
	})
}

// condemnGainLife is Condemn's second sentence: the target's controller
// (carried) gains life equal to its toughness (N). An on-disk key: never
// renamed, never reused.
var condemnGainLife = CardTail("condemn/gain-life", func(ctx *Context, o game.TailOutcome) error {
	player := tailCarry0(o)
	if player == uuid.Nil || o.Tail.N <= 0 {
		return nil
	}
	return GainLife{Player: player, Amount: o.Tail.N}.Apply(ctx)
})
