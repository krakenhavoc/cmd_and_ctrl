package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Icy Reception — Instant {1}{U}:
//
//	"Choose one —
//	 • Counter target creature or legendary spell unless its controller
//	   pays {3}.
//	 • Target creature gets -5/-0 until end of turn."
//
// Mode one asks its spell's controller the CR 118.12 question through
// CounterUnlessPaid, which stops the table while it is open. A
// legendary noncreature spell (a legendary artifact or enchantment) is
// a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "612c7807-be70-4e2f-afd2-3d3c03eddbbe",
		Name:         "Icy Reception",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Counter target creature or legendary spell unless its controller pays {3}.",
				TargetSpell("target creature or legendary spell", Or(Creature(), Legendary()))),
			Mode("Target creature gets -5/-0 until end of turn.",
				TargetCreature("target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				t, ok := ModeTarget(ctx, 0)
				if !ok {
					return nil
				}
				return CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{3}",
					Question: "Icy Reception — pay {3} or your spell is countered?",
				}.Apply(ctx)
			case ctx.HasMode(1):
				t, ok := ModeTarget(ctx, 0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: t.ID, Power: -5, Label: "Icy Reception — -5/-0"}.Apply(ctx)
			}
			return nil
		},
	})
}
