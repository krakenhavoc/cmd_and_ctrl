package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of Lifeforce — Enchantment {2}{G}{G}:
//
//	"If this card is in your opening hand, you may begin the game with
//	 it on the battlefield.
//	 Creature spells can't be countered."
//
// The second line is ADR 0106 §4's battlefield static (#1806), in its
// "any player" form, as Gaea's Herald prints it.
//
// The opening-hand clause (CR 103.6a) is Spec.OpeningHand (ADR 0133):
// the seat holding it is asked as the mulligan window closes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "997478aa-b790-4269-a626-abf0cb30fea0",
		Name:         "Leyline of Lifeforce",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		SpellsCantBeCountered: []game.CounterShieldStatic{
			AnyPlayersSpellsCantBeCountered("Creature spells can't be countered.", Creature()),
		},
	})
}
