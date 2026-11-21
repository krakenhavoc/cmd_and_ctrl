package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turn Inside Out — Instant {R}:
//
//	"Target creature gets +3/+0 until end of turn. When it dies this
//	 turn, manifest dread."
//
// The second sentence is an event-delayed trigger (CR 603.7) pinned to
// the creature, like Whippoorwill's exile: it fires on that object
// going from the battlefield to a graveyard before the turn ends, and
// a flicker ends it (CR 400.7). The manifest dread is the spell
// controller's, not the creature's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be0a3925-8e0d-4ef6-85cb-c9f5eef6b4bb",
		Name:         "Turn Inside Out",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Label: "Turn Inside Out — +3/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			effects.ManifestDreadWhenItDiesThisTurn(ctx, id, "Turn Inside Out — manifest dread")
			return nil
		},
	})
}
