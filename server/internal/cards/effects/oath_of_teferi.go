package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oath of Teferi — Legendary Enchantment {3}{W}{U}:
//
//	"When Oath of Teferi enters, exile another target permanent you
//	 control. Return it to the battlefield under its owner's control at
//	 the beginning of the next end step.
//	 You may activate the loyalty abilities of planeswalkers you control
//	 twice each turn rather than only once."
//
// The enters trigger is Venser, the Sojourner's delayed blink
// (exileTargetsThenScheduleReturn): exiled as it resolves, returned by
// a CR 603.7 delayed trigger at the next end step, under its OWNER's
// control, as a new object. "Another": the Oath cannot target itself.
//
// The static is Spec.YourPlaneswalkersLoyaltyTwiceEachTurn, the class
// form of Urza, Planeswalker's ADR 0145 static: CR 606.3's limit is two
// for every planeswalker its controller controls, read live, so a
// walker that enters later is covered. A second Oath (or Urza) does not
// make it three.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	const name = "Oath of Teferi"
	Register(Spec{
		OracleID:                              "ee00a42d-6ee4-4e46-92d6-62e8e5910880",
		Name:                                  name,
		Completeness:                          CompletenessFull,
		YourPlaneswalkersLoyaltyTwiceEachTurn: true,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(name+" — exile another target permanent you control until the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return exileTargetsThenScheduleReturn(NewContext(g, item), name+" — return the exiled permanent")
					}),
				Another(TargetPermanent("another target permanent you control", YouControl()))),
		},
	})
}
