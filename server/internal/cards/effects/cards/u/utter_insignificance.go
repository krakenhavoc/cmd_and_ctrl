package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Utter Insignificance — Enchantment — Aura {1}{U} (Modern Horizons 3):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature loses all abilities and has base power and
//	 toughness 1/1.
//	 {2}{C}: Exile enchanted creature."
//
// Amphibian Downpour's two statics (the layer 6 removal of every ability,
// then the layer 7b base 1/1, so counters and anthems still count) and Redemption
// Arc's exile of the enchanted creature as an ability of the Aura itself
// (only its controller may activate it; the creature is read as the ability
// resolves, or from the Aura's last-known information if it has left). The
// {C} is colourless mana specifically. The Aura's own ability is not lost
// with the creature's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "44074791-da32-498b-879f-22192b94a9b0",
		Name:            "Utter Insignificance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{C}: Exile enchanted creature.",
			Cost:    effects.ManaCost("{2}{C}"),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.ExileEnchantedCreature,
		}},
	})
}
