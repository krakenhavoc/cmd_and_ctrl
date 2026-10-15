package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restoration Magic — Instant {W} (#3777):
//
//	"Tiered (Choose one additional cost.)
//	 • Cure — {0} — Target permanent gains hexproof and indestructible
//	   until end of turn.
//	 • Cura — {1} — Target permanent gains hexproof and indestructible
//	   until end of turn. You gain 3 life.
//	 • Curaga — {3}{W} — Permanents you control gain hexproof and
//	   indestructible until end of turn. You gain 6 life."
//
// Tiered proof card (CR 702.183a). Curaga's "permanents you control"
// is read once as it resolves (CR 611.2c), so a permanent that enters
// later in the turn gains nothing.
//
// No simplification.
func init() {
	protect := []string{"hexproof", "indestructible"}
	cure := func(life int, label string) func(*game.StackItem, *Context, int) error {
		return func(_ *game.StackItem, ctx *Context, occ int) error {
			if t, ok := ModeTarget(ctx, occ); ok {
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: protect, Label: label}).Apply(ctx); err != nil {
					return err
				}
			}
			if life == 0 {
				return nil
			}
			return GainLife{Player: ctx.Controller(), Amount: life}.Apply(ctx)
		}
	}
	Register(Spec{
		OracleID:     "ddbe0d17-5cac-417c-afc0-e4aebae83205",
		Name:         "Restoration Magic",
		Completeness: CompletenessFull,
		Modes: Tiered(
			SpreeModeDoing("Cure — Target permanent gains hexproof and indestructible until end of turn.", "{0}",
				TargetPermanent("target permanent"), cure(0, "Restoration Magic — Cure")),
			SpreeModeDoing("Cura — Target permanent gains hexproof and indestructible until end of turn. You gain 3 life.", "{1}",
				TargetPermanent("target permanent"), cure(3, "Restoration Magic — Cura")),
			SpreeModeDoing("Curaga — Permanents you control gain hexproof and indestructible until end of turn. You gain 6 life.", "{3}{W}", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := (GrantKeywordUntilEOT{Match: YouControl(), Keywords: protect, Label: "Restoration Magic — Curaga"}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: ctx.Controller(), Amount: 6}.Apply(ctx)
				}),
		),
	})
}
