package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spectral Prison — {1}{U} Aura: "Enchant creature. Enchanted creature
// doesn't untap during its controller's untap step. When enchanted creature
// becomes the target of a spell, sacrifice this Aura."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "b5991a13-5fcd-4d2d-9db2-d3d5af7259f5",
		Name:                  "Spectral Prison",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.AttachedTargetedByASpell,
				"Spectral Prison — sacrifice this Aura", effects.SacrificeThisSource),
		},
	})
}
