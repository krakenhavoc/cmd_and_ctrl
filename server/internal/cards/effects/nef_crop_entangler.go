package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nef-Crop Entangler — Creature — Human Warrior {1}{R}, 2/1:
//
//	"Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	 You may exert this creature as it attacks. When you do, it gets +1/+2 until end of turn. (An exerted creature won't untap during your next untap step.)"
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
		OracleID:        "04dc245b-011f-4207-ab5c-1528037200f5",
		Name:            "Nef-Crop Entangler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		ExertOnAttack:   ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			ExertedGets("Nef-Crop Entangler — it gets +1/+2 until end of turn", 1, 2),
		},
	})
}
