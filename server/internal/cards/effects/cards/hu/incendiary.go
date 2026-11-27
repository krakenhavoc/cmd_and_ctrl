package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incendiary — {R} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 At the beginning of your upkeep, you may put a fuse counter on this Aura.
//	 When enchanted creature dies, this Aura deals X damage to any target,
//	 where X is the number of fuse counters on this Aura."
//
// X is the Aura's counters as it last existed on the battlefield (CR 608.2h),
// and the damage is dealt by the Aura as that object (S82OldB703IncendiaryDamage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a893bee3-ea5c-4619-a875-e4af6d4396fe",
		Name:         "Incendiary",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB703MayPutCounterOnThisAtUpkeep("Incendiary", "fuse"),
			effects.Targeting(
				effects.WhenEnchantedCreatureDies("Incendiary — it deals X damage to any target, where X is the number of fuse counters on it",
					effects.S82OldB703IncendiaryDamage),
				effects.TargetAny()),
		},
	})
}
