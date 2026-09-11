package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snap — Instant for {1}{U}:
//
//	"Return target creature to its owner's hand. Untap up to two
//	 lands."
//
// Free interaction: the two lands pay for the two mana, so Snap is a
// net-zero-mana Unsummon and a storm-deck staple. The untap is the
// whole reason the card is played, so it runs whether or not the
// bounce did anything (it cannot — an illegal target fizzles the
// spell before OnResolve).
//
// Printed order matters and is preserved: bounce first, untap
// second. Bouncing a creature never taps a land, so the order is not
// observable here, but writing it in printed order is how the next
// card that DOES care stays correct.
//
// DECLARED SIMPLIFICATION — THE UNTAP IS NOT A CHOICE. Printed, "up
// to two lands" lets the controller pick any two lands on the
// table, including an opponent's. untapUpToLands takes the first two
// TAPPED LANDS THE CONTROLLER CONTROLS in battlefield order instead
// of opening a picker. Two deviations, both in the weaker direction:
//
//   - An opponent's land is never untapped. Untapping an opponent's
//     land is a real printed option and essentially never a good one;
//     removing it can only cost the controller.
//   - Which two of the controller's own lands is not chosen. That is
//     a real loss when the lands differ (a Volcanic Island is not a
//     Wastes), and it is the reason this is declared rather than
//     waved through.
//
// This is the same call Frantic Search's file makes, in the same
// helper, with the same reasoning stated there: they are your own
// tapped lands and untapping them is strictly good. It stops being
// true exactly when the lands are not interchangeable, which is why
// the note names that case.
func init() {
	Register(Spec{
		OracleID: "ac914d98-221e-426c-8a50-342896b15f9e",
		Name:     "Snap",
		Targets:  TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (BounceToHand{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return untapUpToLands(ctx.Game, item.Controller, 2)
		},
	})
}
