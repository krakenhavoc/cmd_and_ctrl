package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venomous Fangs — Enchantment — Aura {2}{G} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 Whenever enchanted creature deals damage to a creature, destroy the other
//	 creature."
//
// One trigger per damage event, so a creature dealing combat damage to two
// blockers destroys both. Regeneration can save the other creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1ee715e-bab5-4edc-89ba-610cf71bb048",
		Name:         "Venomous Fangs",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.S82OldB725EnchantedDealtDamageToACreature,
				"Venomous Fangs — destroy the other creature", effects.S82OldB725DestroyTheOtherCreature),
		},
	})
}
