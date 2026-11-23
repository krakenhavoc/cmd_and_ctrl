package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "377d31b5-0a48-4d5b-91bd-b5a62d4f7899",
		Name:         "Condemn",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 || targets[0].Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(targets[0].ID)
			if !ok {
				return nil
			}
			return effects.PutIntoLibrary{
				Card:     targets[0].ID,
				ToBottom: true,
				ThenTail: effects.WithN(effects.WithCarry(ctx.Tail(condemnGainLife), c.Controller), c.CurrentToughness()),
			}.Apply(ctx)
		},
	})
}

// condemnGainLife is Condemn's second sentence: the target's controller
// (carried) gains life equal to its toughness (N). An on-disk key: never
// renamed, never reused.
var condemnGainLife = effects.CardTail("condemn/gain-life", func(ctx *effects.Context, o game.TailOutcome) error {
	player := effects.TailCarry0(o)
	if player == uuid.Nil || o.Tail.N <= 0 {
		return nil
	}
	return effects.GainLife{Player: player, Amount: o.Tail.N}.Apply(ctx)
})
