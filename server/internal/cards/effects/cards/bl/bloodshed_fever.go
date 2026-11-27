package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodshed Fever — Enchantment — Aura {R} (Shadowmoor, S82 older sets):
//
//	"Enchant creature
//	 Enchanted creature attacks each combat if able."
//
// Impending Doom's AttacksEachCombatWhere over the attached creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a018a54-6094-42bb-af6a-ef436c09f0b4",
		Name:         "Bloodshed Fever",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.AttacksEachCombatWhere(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == target.InstanceID
			}),
		},
	})
}
