package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patron of the Akki — Legendary Creature — Spirit {4}{R}{R}, 5/5
// (Betrayers of Kamigawa, #5602):
//
//	"Goblin offering (You may cast this spell any time you could cast an
//	 instant by sacrificing a Goblin and paying the difference in mana
//	 costs between this and the sacrificed Goblin. Mana cost includes
//	 color.)
//	 Whenever Patron of the Akki attacks, creatures you control get +2/+0
//	 until end of turn."
//
// Goblin offering is Offering (offering.go, CR 702.48). The attack
// trigger is Dauntless Veteran's: the creatures you control as it
// resolves get the bonus (CR 611.2c locks the set), the Patron among
// them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "38dd86a2-652c-42b7-9b21-2fad059cc706",
		Name:          "Patron of the Akki",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offering("Goblin", Subtype("Goblin"))},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Patron of the Akki — creatures you control get +2/+0 until end of turn",
				Do(BoostUntilEOT{
					Match: And(Creature(), YouControl()), Power: 2,
					Label: "Patron of the Akki — creatures you control get +2/+0",
				})),
		},
	})
}
