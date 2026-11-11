package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// We Ride at Dawn — Enchantment {2}{W} (Outlaws of Thunder Junction
// Commander, #5368):
//
//	"Legendary creature spells you cast have convoke. (Your creatures can
//	 help cast those spells. Each creature you tap while casting a
//	 legendary creature spell pays for {1} or one mana of that creature's
//	 color.)
//	 Whenever your commander attacks, create a 1/1 red Mercenary creature
//	 token with '{T}: Target creature you control gets +1/+0 until end of
//	 turn. Activate only as a sorcery.'"
//
// The convoke line is a granted tap cost (Hoarding Broodlord's shape)
// narrowed to legendary creature spells. The trigger watches a commander
// its controller owns being declared as an attacker; the token is the
// set's Mercenary template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5675ac63-b69b-41b9-a7ed-4085a4e431b2",
		Name:         "We Ride at Dawn",
		Completeness: CompletenessFull,
		GrantedTapCosts: []game.GrantedTapCost{
			SpellsYouCastHaveConvoke("Legendary creature spells you cast have convoke.", SpellIs(And(Creature(), Legendary()))),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, cmdr2S33YourCommanderAttacked,
				"We Ride at Dawn — create a 1/1 red Mercenary", createAMercenary()),
		},
	})
}
