package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hooded Brawler — Creature — Snake Warrior {2}{G}, 3/2:
//
//	"You may exert this creature as it attacks. When you do, it gets +2/+2 until end of turn. (An exerted creature won't untap during your next untap step.)"
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
		OracleID:      "4e40a68a-5e0c-4f4f-9757-02f8139797e2",
		Name:          "Hooded Brawler",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			ExertedGets("Hooded Brawler — it gets +2/+2 until end of turn", 2, 2),
		},
	})
}
