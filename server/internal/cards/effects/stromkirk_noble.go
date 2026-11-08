package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stromkirk Noble — Creature — Vampire Noble {R}, 1/1 (Foundations,
// #4465):
//
//	"This creature can't be blocked by Humans.
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it."
//
// The block rule is Cynical Loner's CantBeBlockedBy with a creature-type
// test (Subtype "Human", read through the layers, so a changeling can't
// block it either). The counter is Hexgold Slith's: on this object only,
// so a Noble that left the battlefield before the trigger resolved gets
// nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c55909d3-1f67-4a4a-9d53-8513d6cf96d8",
		Name:         "Stromkirk Noble",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), Subtype("Human"), "Humans"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Stromkirk Noble — put a +1/+1 counter on it", thisStillHere(plusOneCountersOnThis(1))),
		},
	})
}
