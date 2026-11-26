package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illusionary Armor — Enchantment — Aura {4}{U} (Magic 2014, S82 old sets):
//
//	"Enchant creature
//	 Enchanted creature gets +4/+4.
//	 When enchanted creature becomes the target of a spell or ability, sacrifice
//	 this Aura."
//
// A layer-7c +4/+4, and a trigger on EventBecomesTarget (Fractured Loyalty's
// reading: the engine emits it for a spell or an ability alike, once per
// target, as the targets are chosen) for the creature the Aura is attached to.
// The sacrifice happens as the trigger resolves, above the spell or ability
// that targeted, and does nothing if the Aura has already left.
//
// No simplification.
func init() {
	const label = "Illusionary Armor — sacrifice this Aura"
	effects.Register(effects.Spec{
		OracleID:     "f36603d4-571d-4d03-a9cb-d1f2af3eb3c1",
		Name:         "Illusionary Armor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(4, 4)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesTarget},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID != uuid.Nil && source.IsAttachedTo(ev.CardID)
			},
			Key: label,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, label)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
