package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brass Man — Artifact Creature — Construct {1}, 1/3 (Arabian Nights,
// #5641):
//
//	"This creature doesn't untap during your untap step.
//	 At the beginning of your upkeep, you may pay {1}. If you do, untap
//	 this creature."
//
// Island Fish Jasconius's first two lines at {1}: the catalog's
// untap-step restriction and a "you may pay" on the upkeep trigger's
// resolution (MayPay). The trigger goes on the stack whether or not the
// Brass Man is tapped, as printed; paying untaps it only while it is
// still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "64f56228-7874-4465-ba58-1049083ea02f",
		Name:                  "Brass Man",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Brass Man — you may pay {1} to untap it",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Brass Man — pay {1} to untap it?",
						Then:     brassManPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// brassManPaid is the paid half, keyed (#4264). An on-disk key: never
// renamed, never reused.
var brassManPaid = PayUnlessBranches("pay-unless/brass-man-untap",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if !onBattlefield(ctx.Game, ctx.Item.SourceCardID) {
			return nil
		}
		return UntapTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
	}, nil)
