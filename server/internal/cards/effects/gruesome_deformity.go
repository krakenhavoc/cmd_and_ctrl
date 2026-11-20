package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gruesome Deformity — Enchantment — Aura {B} (Innistrad, #5572):
//
//	"Enchant creature
//	 Enchanted creature has intimidate. (It can't be blocked except by
//	 artifact creatures and/or creatures that share a color with it.)"
//
// Sleeper's Guile's keyword grant with intimidate (GrantToAttached, a
// layer-6 grant) and no other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b44cc9e-6ef0-4fee-b79a-3e34cb0a3407",
		Name:         "Gruesome Deformity",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{GrantToAttached("intimidate")},
	})
}
