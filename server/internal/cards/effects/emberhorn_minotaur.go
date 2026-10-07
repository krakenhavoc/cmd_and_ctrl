package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emberhorn Minotaur — Creature — Minotaur Warrior {3}{R}, 4/3:
//
//	"You may exert this creature as it attacks. When you do, it gets +1/+1 and gains menace until end of turn. (An exerted creature won't untap during your next untap step.)"
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
		OracleID:      "c2d258be-8634-4534-bb09-c064329f0940",
		Name:          "Emberhorn Minotaur",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			ExertedGets("Emberhorn Minotaur — it gets +1/+1 and gains menace until end of turn", 1, 1, "menace"),
		},
	})
}
