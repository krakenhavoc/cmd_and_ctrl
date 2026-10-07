package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watchful Naga — Creature — Snake Wizard {2}{G}, 2/2:
//
//	"You may exert this creature as it attacks. When you do, draw a
//	 card. (An exerted creature won't untap during your next untap
//	 step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h). The draw is unconditional and happens even if
// the Naga has left the battlefield by then.
//
// No simplification.
func init() {
	const label = "Watchful Naga — draw a card"
	Register(Spec{
		OracleID:      "21932be9-cdfb-48d9-87c7-3d62eca98150",
		Name:          "Watchful Naga",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenExerted(label, Do(DrawCards{N: 1})), game.Purpose{Draws: 1}),
		},
	})
}
