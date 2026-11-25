package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skin Invasion // Skin Shedder — Enchantment — Aura {R} // Creature — Insect
// Horror 3/4 (Shadows over Innistrad, S82 OLD):
//
//	Front: "Enchant creature
//	        Enchanted creature attacks each combat if able.
//	        When enchanted creature dies, return this card to the battlefield
//	        transformed under your control."
//	Back:  (no abilities)
//
// Fealty to the Realm's requirement over the attached creature, and Minion's
// Return's death trigger: it fires as the creature dies, before the Aura falls
// to the graveyard, and returns that card back face up only if it is still the
// graveyard object it became (CR 400.7). The Skin Shedder is a creature, so it
// enters unattached.
//
// No simplification.
func init() {
	const oracle = "b5a146e6-053f-43ce-8b01-24505a398976"
	attack := effects.AttacksEachCombatWhere(effects.AttachedToSource)
	attack.Label = "Enchanted creature attacks each combat if able."
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Skin Invasion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{attack},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Skin Invasion — return this card to the battlefield transformed under your control",
				effects.OldB289ReturnThisCardTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Skin Shedder",
		Completeness: effects.CompletenessFull,
	})
}
