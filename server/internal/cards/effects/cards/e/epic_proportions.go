package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Epic Proportions — Enchantment — Aura {4}{G}{G} (Commander 2018, S82 OLD
// batch 193, #5526):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +5/+5 and has trample."
//
// Frantic Strength's statics (PumpAttached, GrantToAttached) on an Aura with
// flash (CR 702.8a), so it can be cast in combat after blockers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad8edfdf-6915-4fc2-9226-3aae95951935",
		Name:            "Epic Proportions",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(5, 5),
			effects.GrantToAttached("trample"),
		},
	})
}
