package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thrive — Sorcery {X}{G} (S82 GAP, #5948):
//
//	"Put a +1/+1 counter on each of X target creatures."
//
// Change of Plans' X-counted targets: X is announced with the spell, and each
// target is re-checked as it resolves (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe3c524e-10b7-485a-9285-973092eeb214",
		Name:         "Thrive",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      effects.TargetsCountedByX(effects.TargetCreature("X target creatures")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
