package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sigil of Sleep — {U} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 Whenever enchanted creature deals damage to a player, return target
//	 creature that player controls to its owner's hand."
//
// Any damage, combat or not. The target is chosen from the creatures the
// damaged player controls (Mistblade Shinobi's clause).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51b47053-d169-4627-982f-f29b8630d75f",
		Name:         "Sigil of Sleep",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDealDamage, effects.S82OldB703AttachedCreatureDealtDamageToPlayer,
					"Sigil of Sleep — return target creature that player controls to its owner's hand",
					effects.BounceUpToOneTarget),
				effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls"))),
		},
	})
}
