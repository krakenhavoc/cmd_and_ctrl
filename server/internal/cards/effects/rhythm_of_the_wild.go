package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhythm of the Wild — Enchantment {1}{R}{G}:
//
//	"Creature spells you control can't be countered.
//	 Nontoken creatures you control have riot. (They enter with your
//	 choice of a +1/+1 counter or haste.)"
//
// Two statics. The first is ADR 0106 §4's battlefield static over the
// counter gate (Prowling Serpopard's second sentence). The second is a
// layer-6 keyword grant, and riot is the engine's (ADR 0109 §10): a
// nontoken creature entering under Rhythm has riot as it would exist on
// the battlefield, so the CR 614.12 look-ahead finds it and the creature
// is asked as it enters — once more than its printed riot, if it has one
// (CR 702.136b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acfa77fd-3610-4f12-9c3c-bd860ce91700",
		Name:         "Rhythm of the Wild",
		Completeness: CompletenessFull,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Creature spells you control can't be countered.", Creature()),
		},
		Static: []game.StaticAbility{NontokenCreaturesYouControlHave(game.KeywordRiot)},
	})
}
