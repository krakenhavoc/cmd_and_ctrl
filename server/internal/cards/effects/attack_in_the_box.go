package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Attack-in-the-Box — Artifact Creature — Toy {3}, 2/4 (Duskmourn: House
// of Horror, #4466):
//
//	"Whenever this creature attacks, you may have it get +4/+0 until end
//	 of turn. If you do, sacrifice it at the beginning of the next end
//	 step."
//
// One optional trigger (CR 603.5): the player is asked as it resolves, and
// a yes pumps the box and schedules the sacrifice, the delayed trigger
// Chrome Dome's token uses. A box that left or came back as a new object in
// the meantime is not "it" (CR 400.7), so nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0fd78bb-cb83-4dfa-a910-0334d225f87c",
		Name:         "Attack-in-the-Box",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WheneverThisAttacks("Attack-in-the-Box — you may have it get +4/+0 until end of turn and sacrifice it at the next end step",
				stdB26AttackInTheBox), "Attack-in-the-Box — get +4/+0 until end of turn and sacrifice it at the beginning of the next end step?"),
		},
	})
}
