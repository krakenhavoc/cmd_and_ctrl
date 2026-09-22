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
// (`TargetSpell(..., HasASingleTarget())` + `ChangeTargets{Policy:
// game.RetargetChangeOne}`, #1206) for the retarget bullet, unblocked
// mid-batch by the "Stack-item retarget" seam closing; and
// RestrictUntilEOT's game.CantBlock bit over one or two announced
// targets for the third.
//
// DECLARED CAVEAT, the same one Bolt Bend and Deflecting Swat ship:
// the retarget bullet reaches SPELLS only. The engine cannot target
// an ability item on the stack at all (ADR 0065's open item, not a
// new one), so "or ability" is narrower than printed — never
// stronger (#259).
func init() {
	Register(Spec{
		OracleID:     "2c49d24a-98a4-43c2-bb59-1ef13d4214c2",
		Name:         "Untimely Malfunction",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The retarget mode can only choose a SPELL, not an activated or triggered ability on the stack — the engine cannot target an ability item."},
		Modes: ChooseOne(
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Change the target of target spell with a single target.",
				TargetSpell("target spell with a single target", HasASingleTarget())),
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
