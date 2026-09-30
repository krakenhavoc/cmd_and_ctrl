package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lure — Enchantment — Aura, {1}{G}{G}:
//
//	"Enchant creature
//	 All creatures able to block enchanted creature do so."
//
// The card CR 509.1c was written around (#1597). The requirement sits
// on the enchanted creature and is one requirement on EACH creature
// able to block it, so the engine refuses a defender's pass while any
// of them is still at home, refuses sending one to block something
// else, and asks nothing of a creature that can't block it — a
// non-flyer against a flyer, a lone blocker when it has menace
// (CR 702.111b).
//
// The static re-reads the attachment on every recompute, so moving or
// destroying the Aura moves or ends the requirement at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a7425ba-4478-4bc4-855f-abf947ea4fa2",
		Name:         "Lure",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			BlockRequirementWhere(game.BlockRequirementLure, AttachedToSource),
		},
	})
}
