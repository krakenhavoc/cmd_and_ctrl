package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viridian Harvest — Enchantment — Aura {G}:
//
//	"Enchant artifact
//	 When enchanted artifact is put into a graveyard, you gain 6 life."
//
// The trigger reads the permanent the Aura was on as it went (the Aura is
// still attached when the event is raised), and fires for any artifact the
// Aura enchants, a creature or not.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48dc8f34-8c73-4155-8a3d-b26a35260e6b",
		Name:         "Viridian Harvest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Artifact()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, LifeGain: 6},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Viridian Harvest — you gain 6 life",
				effects.Do(effects.GainLife{Amount: 6})),
		},
	})
}
