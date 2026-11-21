package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0f0959ff-4464-4738-9d79-5b06f368f144",
		Name:         "Inkling Mascot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.InformedInkwrightCast,
				"Inkling Mascot — Repartee: gains flying until end of turn, surveil 1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := effects.SocB27ThisGetsAndGains(ctx, 0, "flying", "Inkling Mascot"); err != nil {
						return err
					}
					return effects.Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
				}),
		},
	})
}
