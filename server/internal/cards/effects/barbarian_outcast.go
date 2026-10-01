package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barbarian Outcast — Creature {1}{R}, 2/2:
//
//	"When you control no Swamps, sacrifice this creature."
//
// ADR 0107 §1 (#1858). A CR 603.8 state trigger: it triggers as soon
// as its controller controls no permanent with the Swamp subtype, a
// layer-4 grant included (Urborg makes every land a Swamp), and does not
// trigger again while it waits or is on the stack. If a Swamp arrives in
// response the creature is still sacrificed — the condition is not an
// intervening "if" (CR 603.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49d44e89-b94e-43f7-a996-f50651df9264",
		Name:         "Barbarian Outcast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(Subtype("Swamp"), "Barbarian Outcast — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
