package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Link — Enchantment — Aura {1}{W}{B} (Apocalypse, S82 OLD, #5614):
//
//	"Enchant creature
//	 Whenever enchanted creature deals damage, you gain that much life.
//	 Whenever enchanted creature is dealt damage, you gain that much life."
//
// Spirit Link's trigger for the damage the creature deals, and its mirror for
// the damage it is dealt: two printed triggers, two registered ones. "That
// much" is the event's amount after any reduction, and the life goes to the
// Aura's controller (CR 109.5's "you"), whoever controls the creature.
//
// No simplification.
func init() {
	gain := func(g *game.Game, item *game.StackItem) error {
		return effects.GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
	}
	effects.Register(effects.Spec{
		OracleID:     "bb81aa50-49db-4743-85fc-6b4d9ed29bc6",
		Name:         "Soul Link",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Amount > 0 && source.IsAttachedTo(ev.Source)
				},
				Key:    "Soul Link — enchanted creature deals damage: gain that much life",
				Effect: gain,
			},
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
				},
				Key:    "Soul Link — enchanted creature is dealt damage: gain that much life",
				Effect: gain,
			},
		},
	})
}
