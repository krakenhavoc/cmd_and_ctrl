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
// Sandbox simplification, declared: first strike is the point of the
// card — in paper the counters land between the first-strike and
// regular damage steps, so the rest of the team hits harder in the
// same combat. The engine deals both damage steps inside one step
// entry with no priority window between them, so Drana's trigger
// resolves after regular damage has already been dealt. The counters
// still go on every attacker and stay — the team is bigger from the
// next combat on, not this one. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:        "89b24b5f-d837-4274-8877-8ff7dc2708ba",
		Name:            "Drana, Liberator of Malakir",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The +1/+1 counters land after this combat's regular damage rather than between first-strike and regular damage, so the rest of the team doesn't hit harder in the same combat."},
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Drana — a +1/+1 counter on each attacking creature you control", func(g *game.Game, item *game.StackItem) error {
				return b13PutCounterOnEach(NewContext(g, item), b13AttackingCreaturesYouControl(g, item.Controller))
			}),
		},
	})
}
