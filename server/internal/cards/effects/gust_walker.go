package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gust Walker — Creature — Human Wizard {1}{W}, 2/2:
//
//	"You may exert this creature as it attacks. When you do, it gets +1/+1 and gains flying until end of turn. (An exerted creature won't untap during your next untap step.)"
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
		OracleID:      "48bf69ec-0030-44ee-9fbd-6c204b64da7d",
		Name:          "Gust Walker",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			ExertedGets("Gust Walker — it gets +1/+1 and gains flying until end of turn", 1, 1, "flying"),
		},
	})
}
