package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhonas's Stalwart — {1}{G} Creature — Human Warrior 2/2:
//
//	"You may exert this creature as it attacks. When you do, it gets
//	 +1/+1 until end of turn and can't be blocked by creatures with
//	 power 2 or less this turn. (An exerted creature won't untap during
//	 your next untap step.)"
//
// ADR 0130 §11 (exert as it attacks, CR 701.43d) with a linked "when
// you do" (CR 607.2h). The pump and the block rule are one scoped
// record, put on the stack with the attack so both are in place before
// blockers are declared. The rule is cantBeBlockedByPower (#2600): the
// blocker's power is read live when blockers are declared (CR 509.1b),
// pinned to this creature at resolution (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8c95230b-7a4a-4293-ab5f-2070d951082b",
		Name:          "Rhonas's Stalwart",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			ExertedGetsAndCantBeBlockedByPower("Rhonas's Stalwart — it gets +1/+1 and can't be blocked by creatures with power 2 or less until end of turn", 1, 1, 2),
		},
	})
}
