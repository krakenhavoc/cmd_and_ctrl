package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insidious Will — Instant {2}{U}{U}:
//
//	"Choose one —
//	 • Counter target spell.
//	 • You may choose new targets for target spell.
//	 • Copy target instant or sorcery spell. You may choose new targets
//	   for the copy."
//
// Each bullet is one existing primitive over its own clause: the
// counter, the CR 115.7c "you may" retarget Redirect uses, and the
// Twincast copy. The mode is chosen at announce and only the chosen
// bullet's target is picked (CR 700.2), so the bullets do not share a
// clause. The retarget bullet has no single-target restriction.
func init() {
	Register(Spec{
		OracleID:     "8088d110-2314-45c3-b6ee-3f55d1f388e4",
		Name:         "Insidious Will",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Counter target spell.", TargetSpell("target spell"), CounterTheModesTarget),
			ModeDoing("You may choose new targets for target spell.", TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ChangeTargets{
						StackID:  t.ID,
						Policy:   game.RetargetChooseNew,
						Optional: true,
						Reason:   "Insidious Will — choose new targets",
					}.Apply(ctx)
				}),
			ModeDoing("Copy target instant or sorcery spell. You may choose new targets for the copy.",
				instantOrSorcerySpell("target instant or sorcery spell"),
				copyTheModesSpellOrAbility),
		),
	})
}
