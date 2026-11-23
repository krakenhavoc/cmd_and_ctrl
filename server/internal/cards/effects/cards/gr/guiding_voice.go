package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guiding Voice — Sorcery {W} (EDHREC rank 23915):
//
//	"Put a +1/+1 counter on target creature.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// Learn is the last instruction (CR 701.48a, learn.go). A creature that
// left the battlefield in response gets no counter (CR 608.2b); the
// spell, whose only target is gone, is countered whole before any of
// this runs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c7e7c9e-00d1-497e-bee7-278bbd82e234",
		Name:         "Guiding Voice",
		Purpose:      game.Purpose{Learn: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.Learn{},
		),
	})
}
