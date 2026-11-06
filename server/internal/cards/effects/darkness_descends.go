package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Darkness Descends — Sorcery {2}{B}{B} (Lorwyn Eclipsed, #4458):
//
//	"Put two -1/-1 counters on each creature."
//
// Every creature on the battlefield, each player's, is collected as the
// spell resolves (CR 608.2) and gets its two counters as one placement,
// all of them together: a counter replacement applies once per creature
// and a "one or more counters" watcher sees each creature once. The
// creatures it shrinks to 0 toughness die afterwards, as a state-based
// action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3435757-d049-4aaf-a2cb-897dd4940997",
		Name:         "Darkness Descends",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var creatures []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() {
					creatures = append(creatures, c.InstanceID)
				}
			}
			for _, id := range creatures {
				if err := (AddCounter{Target: id, Kind: game.CounterMinusOne, N: 2}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
