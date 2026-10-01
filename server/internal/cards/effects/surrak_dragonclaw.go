package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surrak Dragonclaw — Legendary Creature — Human Warrior {2}{G}{U}{R}, 6/6:
//
//	"Flash
//	 This spell can't be countered.
//	 Creature spells you control can't be countered.
//	 Other creatures you control have trample."
//
// Flash is a keyword; the rider is Spec.CantBeCountered; the third line
// is ADR 0106 §4's battlefield static (#1806); the last is an ordinary
// layer-6 keyword grant over other creatures Surrak's controller
// controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c766365-82c2-46d6-8521-42e73129f4ef",
		Name:            "Surrak Dragonclaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CantBeCountered: true,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Creature spells you control can't be countered.", Creature()),
		},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, "trample"),
		},
	})
}
