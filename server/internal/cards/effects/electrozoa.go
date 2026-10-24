package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

var electrozoaTap = energyDeclined("pay-unless/electrozoa-tap",
	func(g *game.Game, item *game.StackItem) error {
		if !sourceIsStillThisPermanent(g, item) {
			return nil
		}
		return TapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
	})

// Electrozoa — Creature — Jellyfish {2}{U}, 3/1:
//
//	"Flash
//	 Flying
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 At the beginning of your first main phase, tap this creature unless
//	 you pay {E}."
//
// ADR 0129 §3 (#1995): "tap it unless you pay {E}" is CR 118.12a's
// pay-unless with an energy payment; declining, or being short (CR
// 118.3), taps the Jellyfish if it is still the creature that
// triggered (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "202091b2-9e08-47c1-9099-628b21d8ac12",
		Name:            "Electrozoa",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Electrozoa", 2),
			AtYourPrecombatMain("Electrozoa — tap it unless you pay {E}", func(g *game.Game, item *game.StackItem) error {
				return PayEnergyUnless{
					N:        1,
					Question: "Electrozoa — pay {E}, or tap it?",
					Else:     electrozoaTap,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
