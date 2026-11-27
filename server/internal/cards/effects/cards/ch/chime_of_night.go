package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chime of Night — {1}{B} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 When this Aura is put into a graveyard from the battlefield, destroy
//	 target nonblack creature."
//
// Brilliant Halo's leaves-the-battlefield trigger, however the Aura got there,
// with a targeted destroy (regeneration allowed).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09b41783-0583-4486-99e8-3f8a11ae37fd",
		Name:         "Chime of Night",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Chime of Night — destroy target nonblack creature",
					effects.Do(effects.Instruction(effects.DestroyTheFirstLegalCardTarget))),
				effects.TargetCreature("target nonblack creature", effects.NonBlack())),
		},
	})
}
