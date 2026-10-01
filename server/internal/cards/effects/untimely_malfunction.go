package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Untimely Malfunction — Instant {1}{R} (EDHREC rank 265):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Change the target of target spell or ability with a single
//	   target.
//	 • One or two target creatures can't block this turn."
//
// All three bullets, on the #764 modal machinery: a plain
// DestroyTarget; Bolt Bend's exact CR 115.7b shape
// (`TargetSpellOrAbility(..., ItemHasASingleTarget())` +
// `ChangeTargets{Policy: game.RetargetChangeOne}`) for the retarget
// bullet, which reaches an ability on the stack since #1211; and
// RestrictUntilEOT's game.CantBlock bit over one or two announced
// targets for the third.
func init() {
	Register(Spec{
		OracleID:     "2c49d24a-98a4-43c2-bb59-1ef13d4214c2",
		Name:         "Untimely Malfunction",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Change the target of target spell or ability with a single target.",
				TargetSpellOrAbility("target spell or ability with a single target", ItemHasASingleTarget())),
			Mode("One or two target creatures can't block this turn.",
				TargetCreature("one or two target creatures").WithCount(1, 2)),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				for _, t := range ctx.ModeTargets(0) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			case ctx.HasMode(1):
				for _, t := range ctx.ModeTargets(0) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (ChangeTargets{
						StackID: t.ID,
						Policy:  game.RetargetChangeOne,
						Reason:  "Untimely Malfunction — change the target",
					}).Apply(ctx); err != nil {
						return err
					}
				}
			case ctx.HasMode(2):
				for _, t := range ctx.ModeTargets(0) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBlock,
						Label:        "Untimely Malfunction — can't block this turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
