package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inkling Mascot — Creature — Inkling Cat {W}{B}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"Repartee — Whenever you cast an instant or sorcery spell that targets a
//	 creature, this creature gains flying until end of turn. Surveil 1."
//
// Repartee is Informed Inkwright's trigger (informedInkwrightCast: the cast
// spell's own announced targets). The flying belongs to this permanent and is
// skipped if it has left or come back as a new object (CR 400.7); the surveil
// is the controller's and happens regardless.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f0959ff-4464-4738-9d79-5b06f368f144",
		Name:         "Inkling Mascot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, informedInkwrightCast,
				"Inkling Mascot — Repartee: gains flying until end of turn, surveil 1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := socB27ThisGetsAndGains(ctx, 0, "flying", "Inkling Mascot"); err != nil {
						return err
					}
					return Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
				}),
		},
	})
}
