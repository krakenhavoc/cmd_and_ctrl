package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rain of Daggers — Sorcery {4}{B}{B} (Mystery Booster 2, #5948):
//
//	"Destroy all creatures target opponent controls. You lose 2 life for
//	 each creature destroyed this way."
//
// Overwhelming Forces' wipe with a life payment in place of the draw: the
// loss counts the creatures really destroyed, so an indestructible creature
// or one saved by a regeneration shield costs nothing (CR 701.8, 702.12b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2048201-6dc9-4cf5-916f-1d867ae8dbdd",
		Name:         "Rain of Daggers",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true},
			Answers: game.AnswerRemove,
		},
		Targets: TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opp := TargetedPlayer(ctx)
			if opp == uuid.Nil {
				return nil
			}
			return DestroyAllMatching{
				Match:    And(Creature(), ControlledBy(opp)),
				ThenTail: ctx.Tail(rainOfDaggersLoseLife),
			}.Apply(ctx)
		},
	})
}

// rainOfDaggersLoseLife is "you lose 2 life for each creature destroyed this
// way". An on-disk key: never renamed, never reused.
var rainOfDaggersLoseLife = MassTail("mass/rain-of-daggers-lose-life", func(ctx *Context, _ []game.Card, destroyed int) error {
	if destroyed <= 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2*destroyed)
})
