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
// Two of the three bullets, on the #764 modal machinery: a plain
// DestroyTarget, and RestrictUntilEOT's game.CantBlock bit over one
// or two announced targets (RogueUE's Passage's turn-scoped
// restriction primitive, widened to a variable count).
//
// DECLARED SIMPLIFICATION, weaker than printed: the retarget bullet
// is not offered. Rewriting a stack item's target at resolution
// (docs/engine-seams.md, "Stack-item retarget") has no `*ForEffect`
// primitive yet; #764's modal machinery lets the OTHER two bullets
// stand on their own regardless.
func init() {
	Register(Spec{
		OracleID:     "2c49d24a-98a4-43c2-bb59-1ef13d4214c2",
		Name:         "Untimely Malfunction",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only two of the three modes are offered — \"change the target of target spell or ability\" isn't implemented."},
		Modes: ChooseOne(
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("One or two target creatures can't block this turn.",
				TargetCreature("one or two target creatures").WithCount(1, 2)),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				for _, t := range ctx.ModeTargets(0) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}
			if ctx.HasMode(1) {
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
