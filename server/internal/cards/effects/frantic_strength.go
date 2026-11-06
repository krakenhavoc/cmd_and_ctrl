package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frantic Strength — Enchantment — Aura {2}{G} (Duskmourn: House of
// Horror, #4466):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +2/+2 and has trample."
//
// Etali's Favor's statics (PumpAttached, GrantToAttached) on an Aura with
// flash (CR 702.8a), so it can be cast in combat after blockers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5dbe4648-f4c9-4d26-98f5-58a7e53b149d",
		Name:            "Frantic Strength",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("trample"),
		},
	})
}
