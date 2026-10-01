package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hexing Squelcher — Creature — Goblin Sorcerer {1}{R}, 2/2:
//
//	"This spell can't be countered.
//	 Ward—Pay 2 life.
//	 Spells you control can't be countered.
//	 Other creatures you control have "Ward—Pay 2 life.""
//
// The rider is Spec.CantBeCountered; the third line is ADR 0106 §4's
// battlefield static (#1806). Both wards are ward.go's: the printed one
// is Ward, and the granted one is WardGranted over the other creatures
// the Squelcher's controller controls, so the same trigger body and the
// same pay-2-life prompt serve both (Refraction Elemental's cost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "96c2d71e-0fc9-42aa-bc6b-6d0ae3f66f8b",
		Name:            "Hexing Squelcher",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Spells you control can't be countered."),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(2), "Hexing Squelcher — ward, pay 2 life"),
			WardGranted(WardLife(2), "Hexing Squelcher — other creatures you control have ward, pay 2 life",
				TribeFilter{Others: true, YoursOnly: true}.Matches),
		},
	})
}
