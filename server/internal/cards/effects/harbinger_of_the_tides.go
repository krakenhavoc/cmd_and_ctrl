package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harbinger of the Tides — Creature — Merfolk Wizard {U}{U}, 2/2
// (Foundations, #4465):
//
//	"You may cast this spell as though it had flash if you pay {2} more
//	 to cast it. (You may cast it any time you could cast an instant.)
//	 When this creature enters, you may return target tapped creature an
//	 opponent controls to its owner's hand."
//
// The surcharge is Asinine Antics's FlashForMore offer ({U}{U} plus {2}),
// so it carries that card's caveat (flash_for_more.go, #5207). The entry
// trigger is Spellweaver Duo's optional bounce (CR 603.5) with the target
// narrowed to a creature an opponent controls; whether it is tapped is
// read as the target is chosen and again on resolution (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "9f5277e8-1cbe-478c-a8bc-ac028f13305e",
		Name:         "Harbinger of the Tides",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Paying {2} more to cast this at instant speed can't be combined with another alternative cost.",
		},
		AlternativeCosts: []game.AlternativeCost{FlashForMore("{2}{U}{U}", "{2}")},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Harbinger of the Tides — return target tapped creature an opponent controls to its owner's hand", bounceTheTargetToHand),
				TargetCreature("target tapped creature an opponent controls", hobB3TappedCreature(), OpponentControls())),
				"Harbinger of the Tides — return target tapped creature an opponent controls to its owner's hand?"),
		},
	})
}
