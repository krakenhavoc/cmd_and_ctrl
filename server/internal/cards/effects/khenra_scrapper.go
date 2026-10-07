package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Khenra Scrapper — Creature — Jackal Warrior {2}{R}, 2/3:
//
//	"Menace (A creature with menace can't be blocked except by two or more creatures.)
//	 You may exert this creature as it attacks. When you do, it gets +2/+0 until end of turn. (An exerted creature won't untap during your next untap step.)"
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
		OracleID:        "6ab8958b-10a2-4601-a609-7a372a47f36c",
		Name:            "Khenra Scrapper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		ExertOnAttack:   ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			ExertedGets("Khenra Scrapper — it gets +2/+0 until end of turn", 2, 0),
		},
	})
}
