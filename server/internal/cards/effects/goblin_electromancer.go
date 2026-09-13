package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Electromancer — Creature — Goblin Wizard {U}{R}, 2/2:
//
//	"Instant and sorcery spells you cast cost {1} less to cast."
//
// The first reducer in the catalog, and the card that pins CR
// 601.2f's ordering: a Lightning Bolt under a Sphere of Resistance
// AND an Electromancer costs {R}, not {1}{R}, because the increase
// resolves first and leaves a generic symbol for the reduction to
// spend. Run the passes the other way round and the reduction finds
// nothing to reduce.
//
// "YOU cast" — YourSpell() compares the caster against the
// Electromancer's controller, so an opponent's Counterspell is
// untouched. And the reduction spends generic mana only, so a {U}{U}
// Counterspell stays {U}{U}: this is the natural card on which to
// notice that rule.
func init() {
	Register(Spec{
		OracleID: "81f06f84-1580-43c0-89d5-08d34541a519",
		Name:     "Goblin Electromancer",
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
	})
}
