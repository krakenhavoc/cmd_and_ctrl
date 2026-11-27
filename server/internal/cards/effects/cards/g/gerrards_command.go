package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gerrard's Command — Instant {G}{W} (Planeshift, S82 OLD #5615):
//
//	"Untap target creature. It gets +3/+3 until end of turn."
//
// Spidery Grasp's two steps in order: the untap, then the boost, both on the
// target if it is still legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de8ebe4b-d86d-478b-946e-f83e1409b63f",
		Name:         "Gerrard's Command",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Gerrard's Command — +3/+3"}.Apply(ctx)
		},
	})
}
