package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrull Retainer — Enchantment — Aura {B} (Fallen Empires, S82 OLD b855, #5637):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 Sacrifice this Aura: Regenerate enchanted creature."
//
// Briar Shield's shape with a regeneration for the sacrifice: the shield goes to the
// creature the Aura was attached to as it left (CR 608.2h), if it is still on the
// battlefield, so it is made after the Aura's +1/+1 is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81a47019-1343-4a63-893a-0d8ba5e848d5",
		Name:         "Thrull Retainer",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this Aura: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.S82OldB855RegenerateEnchantedCreature,
		}},
	})
}
