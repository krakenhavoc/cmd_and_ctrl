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
// ONE SIMPLIFICATION, the one every Leyline in the catalog carries
// (leyline_of_sanctity.go): there is no pre-game window in which a card
// in an opening hand can be put onto the battlefield, so the Leyline is
// cast for its four mana like any other enchantment. Strictly weaker
// than printed.
func init() {
	Register(Spec{
		OracleID:     "997478aa-b790-4269-a626-abf0cb30fea0",
		Name:         "Leyline of Lifeforce",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Starting the game with this on the battlefield from your opening hand isn't implemented — you cast it for {2}{G}{G} like an ordinary enchantment.",
		},
		SpellsCantBeCountered: []game.CounterShieldStatic{
			AnyPlayersSpellsCantBeCountered("Creature spells can't be countered.", Creature()),
		},
	})
}
