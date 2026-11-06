package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Might of the Ancestors — Enchantment {2}{W} (The Lost Caverns of
// Ixalan, #4470):
//
//	"At the beginning of combat on your turn, target creature you control
//	 gets +2/+0 and gains vigilance until end of turn."
//
// Brambleguard Captain's trigger shape (AtBeginningOfYourCombat with a
// target creature you control) with the shared pump-and-grant body: the
// still-legal target (CR 608.2b) gets +2/+0 and vigilance (CR 702.20)
// until end of turn, and the grant lands before attackers are declared.
// With no creature to target, the trigger is removed from the stack (CR
// 603.3d).
//
// No simplification.
func init() {
	const label = "Might of the Ancestors — target creature you control gets +2/+0 and gains vigilance until end of turn"
	Register(Spec{
		OracleID:     "5e4107ce-8450-4501-9cb3-a6344ef320f1",
		Name:         "Might of the Ancestors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(label, pumpAndGrantFirstTarget(2, 0, label, "vigilance")),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
