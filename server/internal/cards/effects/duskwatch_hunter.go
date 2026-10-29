package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duskwatch Hunter — Creature — Wolf {2}{B/G}, 3/1 (The Hobbit, HOB):
//
//	"This creature can't be blocked by tokens.
//	 When this creature enters, put a +1/+1 counter on target creature."
//
// The evasion is a block rule on the Hunter itself, Legolas Greenleaf's
// shape with a token test for the forbidden blocker; the counter is Hunting
// Moa's (any creature, the Hunter included, chosen as the trigger goes on
// the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6283e5ca-9924-478b-89a5-36924accb9aa",
		Name:         "Duskwatch Hunter",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), IsTokenPredicate(), "tokens"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Duskwatch Hunter — put a +1/+1 counter on target creature", putPlusOneCounterOnEachLegalTarget),
				TargetCreature("target creature")),
		},
	})
}
