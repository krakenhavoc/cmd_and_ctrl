package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lay Bare — Instant {2}{U}{U} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Counter target spell. Look at its controller's hand."
//
// The spell's controller is read before the counter, while it is still on the
// stack (last-known information, CR 608.2h), and only the caster learns the
// hand: a look, not a reveal. A spell that was countered by something else in
// response leaves the target illegal and Lay Bare does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d3ac460-79ac-4eef-9446-1b9fe0261556",
		Name:         "Lay Bare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				owner, _ := ctx.Game.LastKnownSpellControllerForEffect(t.ID)
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				ctx.Game.LookAtHandForEffect(ctx.Controller(), owner)
				return nil
			}
			return nil
		},
	})
}
