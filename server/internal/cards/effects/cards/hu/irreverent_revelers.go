package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	t := effects.WhenThisEnters("Irreverent Revelers — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Destroy target artifact.",
			effects.TargetPermanent("target artifact", effects.Artifact()),
			effects.DestroyTheModesTarget),
		effects.ModeDoing("This creature gains haste until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.SourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"haste"},
					Label:    "Irreverent Revelers — haste until end of turn",
				}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "05abc57a-839d-4376-986c-8e3855e34239",
		Name:         "Irreverent Revelers",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
