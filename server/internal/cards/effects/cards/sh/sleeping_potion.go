package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sleeping Potion — Enchantment — Aura {1}{U} (Planeshift, #5615):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 When enchanted creature becomes the target of a spell or ability,
//	 sacrifice this Aura."
//
// Bitter Chill's tap and untap lock, with Illusionary Armor's sacrifice on a
// target (the engine emits EventBecomesTarget for a spell or an ability alike,
// once per target): the Aura is sacrificed as the trigger resolves, above the
// spell or ability that targeted, and does nothing if it has already left.
//
// No simplification.
func init() {
	const label = "Sleeping Potion — sacrifice this Aura"
	effects.Register(effects.Spec{
		OracleID:              "c8fbcef0-2cbf-4a59-8322-374dd9153ebf",
		Name:                  "Sleeping Potion",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sleeping Potion — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
			{
				Watches: []game.EventKind{game.EventBecomesTarget},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventBecomesTarget && ev.CardID != uuid.Nil && source.IsAttachedTo(ev.CardID)
				},
				Key: label,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, label)
				},
				Effect: effects.S82OldB636SacrificeThisAura,
			},
		},
	})
}
