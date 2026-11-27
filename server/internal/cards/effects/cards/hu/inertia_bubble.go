package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inertia Bubble — Enchantment — Aura {1}{U} (Mirrodin, #5606):
//
//	"Enchant artifact
//	 Enchanted artifact doesn't untap during its controller's untap step."
//
// Bitter Chill's untap restriction (CR 502.3) on an artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "864c7575-4589-416c-b143-310d5ef238c5",
		Name:                  "Inertia Bubble",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.TargetPermanent("enchant artifact", effects.Artifact()),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
