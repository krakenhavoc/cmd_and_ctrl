package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drana, Liberator of Malakir — Legendary Creature — Vampire Ally
// {1}{B}{B}, 2/3 (EDHREC rank 1522):
//
//	"Flying, first strike
//	 Whenever Drana deals combat damage to a player, put a +1/+1
//	 counter on each attacking creature you control."
//
// The aggro deck's team pump. The trigger is the "this creature
// deals combat damage to a player" reader; the set is every creature
// the controller controls that is attacking when the trigger
// RESOLVES (AttackingTarget is stamped at declaration and cleared at
// end of combat), Drana herself included, as printed. Counters go on
// through AddCounter, so a doubler applies.
//
// First strike is the point of the card, and it works as printed: the
// first-strike damage step is a real step with a priority window
// (CR 510.4, #717), and the cursor cannot leave it owing what is on the
// stack (CR 117.4, #914), so the trigger resolves and the counters land
// BEFORE the regular damage step, which reads each creature's power
// then. TestB13DranaGrowsEveryAttackerWhenSheConnects proves it (2 from
// Drana, then a Bear grown to 3). The caveat this card carried said the
// counters came after regular damage; that stopped being true when the
// first-strike step landed. No simplification.
func init() {
	Register(Spec{
		OracleID:        "89b24b5f-d837-4274-8877-8ff7dc2708ba",
		Name:            "Drana, Liberator of Malakir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Drana — a +1/+1 counter on each attacking creature you control", func(g *game.Game, item *game.StackItem) error {
				return b13PutCounterOnEach(NewContext(g, item), b13AttackingCreaturesYouControl(g, item.Controller))
			}),
		},
	})
}
