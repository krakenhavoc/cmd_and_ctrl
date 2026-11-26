package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psychic Strike — Instant {1}{U}{B} (Gatecrash, #5565):
//
//	"Counter target spell. Its controller mills two cards."
//
// Glen Elendra Guardian's shape: the spell's controller is read before the
// counter, and mills whether or not the spell was actually countered (a spell
// that can't be countered still named a legal target). With the target gone
// by resolution the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17694111-5623-48b3-8c54-c383418e8d2b",
		Name:         "Psychic Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				spell := ctx.Game.StackItemForEffect(t.ID)
				if spell == nil {
					return nil
				}
				victim := spell.Controller
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.MillCards{Player: victim, N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
