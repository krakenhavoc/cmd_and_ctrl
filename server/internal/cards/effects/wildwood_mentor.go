package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildwood Mentor — Creature — Treefolk {2}{G}, 1/1 (Wilds of Eldraine,
// #4471):
//
//	"Whenever a token you control enters, put a +1/+1 counter on this
//	 creature.
//	 Whenever this creature attacks, another target attacking creature
//	 gets +X/+X until end of turn, where X is this creature's power."
//
// The first trigger is Belladonna Took's "a token you control enters"
// (any token, not only a creature), and the counter goes on the Mentor
// only while it is still the same permanent. The attack trigger targets
// another attacking creature as it goes on the stack (Restless Ridgeline's
// clause). X is the Mentor's power as the trigger resolves, or its last
// known power if it has left (CR 608.2h, Firebending Student's
// sourcePowerNow); a negative power gives +0/+0 (CR 107.1b). A target that
// is gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Wildwood Mentor — another target attacking creature gets +X/+X until end of turn, where X is its power"
	attack := WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		x := max(sourcePowerNow(ctx), 0)
		if x == 0 {
			return nil
		}
		return BoostUntilEOT{Target: target, Power: x, Toughness: x, Label: "Wildwood Mentor — +X/+X until end of turn"}.Apply(ctx)
	})
	attack.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	Register(Spec{
		OracleID:     "f4e61b03-e6c8-4751-9720-b91fdecd57da",
		Name:         "Wildwood Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, belladonnaTokenYouControlEntered,
				"Wildwood Mentor — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			attack,
		},
	})
}
