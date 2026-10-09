package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Violent Echoes — Instant {2}{R}{R} (Reality Fracture, tracker #2795):
//
//	"Violent Echoes deals 6 damage to target creature or planeswalker. If
//	 excess damage was dealt to that permanent this way, empower Jace X,
//	 where X is that excess damage."
//
// Excess damage (CR 120.4a) is what landed beyond lethal: toughness less
// damage already marked for a creature, the loyalty for a planeswalker
// (frDealDamageWithExcess), read after the damage so prevention lowers
// it. With no excess the keyword action does not happen at all (the
// "if"), so no empty Jace token appears. A target that left is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8ecec61-cc47-4051-a599-0e7470764488",
		Name:         "Violent Echoes",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 6)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				excess, err := frDealDamageWithExcess(ctx, t.ID, 6)
				if err != nil || excess <= 0 {
					return err
				}
				return EmpowerJace{N: excess}.Apply(ctx)
			}
			return nil
		},
	})
}
