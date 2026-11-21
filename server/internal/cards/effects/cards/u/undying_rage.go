package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undying Rage — Enchantment — Aura {2}{R} (Dominaria Remastered, #5948):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and can't block.
//	 When this Aura is put into a graveyard from the battlefield, return it
//	 to its owner's hand."
//
// The pump is PumpAttached, "can't block" is the attached-creature
// restriction Sluggishness uses, and the return is the shared Aura trigger
// (UlgB1AuraReturnsToHand): it fires when the Aura goes from the battlefield to
// a graveyard, for any reason, and returns it to its owner's hand if it is
// still there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "622d783d-4201-4769-8a0f-442b358d8bf1",
		Name:         "Undying Rage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.RestrictAttached(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			effects.UlgB1AuraReturnsToHand("Undying Rage"),
		},
	})
}
