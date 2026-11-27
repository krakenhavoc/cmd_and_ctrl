package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scatter Arc — Instant {3}{U} (Gatecrash, #5565):
//
//	"Counter target noncreature spell.
//	 Draw a card."
//
// The counter and the draw are two instructions of one resolution: the
// draw is the caster's and happens whether or not the spell was actually
// countered (a spell that can't be countered still named a legal target).
// With the target gone by resolution the whole spell is countered on
// resolution for want of a legal target (CR 608.2b) and nothing is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "424683a5-d239-400d-af2b-9c00d1d32d7f",
		Name:         "Scatter Arc",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
