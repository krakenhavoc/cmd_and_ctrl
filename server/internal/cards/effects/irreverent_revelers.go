package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Irreverent Revelers — Creature — Satyr {2}{R}, 2/2 (Theros Beyond Death, #5514):
//
//	"When this creature enters, choose one —
//	 • Destroy target artifact.
//	 • This creature gains haste until end of turn."
//
// Ravenous Sailback's modal enters trigger: the mode is chosen as the
// trigger is put on the stack (CR 700.2a), and the haste mode does
// nothing once the Revelers have left the battlefield.
//
// No simplification.
func init() {
	t := WhenThisEnters("Irreverent Revelers — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Destroy target artifact.",
			TargetPermanent("target artifact", Artifact()),
			DestroyTheModesTarget),
		ModeDoing("This creature gains haste until end of turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if !sourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"haste"},
					Label:    "Irreverent Revelers — haste until end of turn",
				}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "05abc57a-839d-4376-986c-8e3855e34239",
		Name:         "Irreverent Revelers",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
