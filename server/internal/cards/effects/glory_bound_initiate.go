package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glory-Bound Initiate — Creature — Human Warrior {1}{W}, 3/1:
//
//	"You may exert this creature as it attacks. When you do, it gets +1/+3 and gains lifelink until end of turn. (An exerted creature won't untap during your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h) that pumps this creature until end of turn. The
// pump goes on the stack with the attack, so it is in place before
// blockers are declared, and the pump is declared as the row's Purpose
// so the bot prices the exert by it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "b5582be1-6c2f-4052-80da-bf8f1913567b",
		Name:          "Glory-Bound Initiate",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			ExertedGets("Glory-Bound Initiate — it gets +1/+3 and gains lifelink until end of turn", 1, 3, "lifelink"),
		},
	})
}
