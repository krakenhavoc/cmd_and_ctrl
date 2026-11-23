package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mire Blight — Enchantment — Aura {B} (Jumpstart 2022):
//
//	"Enchant creature
//	 When enchanted creature is dealt damage, destroy it."
//
// Pain for All's damage watcher on the enchanted creature (any positive
// damage, combat or not). The creature is named as the trigger goes on the
// stack, so "it" still names the host if the Aura is gone by resolution;
// regeneration and indestructible apply as for any destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcfbe3b8-a955-45d8-9329-3af565a3b53c",
		Name:         "Mire Blight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.EnchantCreature(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
			},
			Key: "Mire Blight — destroy enchanted creature",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Mire Blight — destroy enchanted creature")
				epoch := -1
				if c, ok := g.LookupCardForEffect(ev.Target); ok {
					epoch = c.ObjectEpoch
				}
				item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: epoch}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				c, ok := g.LookupCardForEffect(item.Params.Object.ID)
				if !ok || c.ObjectEpoch != item.Params.Object.Epoch || !effects.OnBattlefield(g, c.InstanceID) {
					return nil
				}
				return effects.DestroyTarget{Target: c.InstanceID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
