package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Energy Arc — Instant {W}{U}:
//
//	"Untap any number of target creatures. Prevent all combat damage that
//	 would be dealt to and dealt by those creatures this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): every legal target untaps, then
// ONE to-and-by record (Mod.AndDealtBy) protects all of them
// (ShieldTheTargetPermanents): combat damage between two of those
// creatures meets one prevention effect, not two (CR 614.5).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "959bcf99-3c9a-4f99-98b2-13514d3dac16",
		Name:         "Energy Arc",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return effects.ToAndByShield(effects.ShieldTheTargetPermanents, true).Apply(ctx)
		},
	})
}
