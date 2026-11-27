package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Momentum — {2}{G} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 At the beginning of your upkeep, you may put a growth counter on this
//	 Aura.
//	 Enchanted creature gets +1/+1 for each growth counter on this Aura."
//
// Banshee's Blade's counted pump (PumpAttachedPer), re-read every layer pass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "80875914-df4c-4b05-8c4d-658ff477c956",
		Name:         "Momentum",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int { return source.Counters["growth"] }),
		},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB703MayPutCounterOnThisAtUpkeep("Momentum", "growth"),
		},
	})
}
