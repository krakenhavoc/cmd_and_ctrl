package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basilisk Collar — Artifact — Equipment for {1} (EDHREC rank 253,
// the highest-ranked Equipment in the catalog after the Boots and the
// Greaves):
//
//	"Equipped creature has deathtouch and lifelink.
//	 Equip {2}"
//
// One mana, no stat line, and two keywords that between them make any
// creature a removal spell that also stabilises. Both are honoured:
// deathtouch is read by the combat damage / lethal check and lifelink
// by the damage application, so a 1/1 wearing the Collar really does
// trade with a 9/9 and gain its controller a life doing it.
//
// The Collar is in this batch for what it does NOT need: no pump, no
// trigger, no target clause beyond the equip. It is GrantToAttached
// and nothing else, which is the shape roughly half the remaining
// Equipment in the roadmap's attachment seam has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5f4dd28-f4ae-4d39-b9b8-6ebfd63c93fe",
		Name:         "Basilisk Collar",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("deathtouch", "lifelink"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
