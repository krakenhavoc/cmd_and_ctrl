package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiteful Shadows — Enchantment — Aura {1}{B} (Dark Ascension, S82 OLD,
// #5571):
//
//	"Enchant creature
//	 Whenever enchanted creature is dealt damage, it deals that much damage
//	 to its controller."
//
// Pain for All's reflect trigger aimed at the creature's controller: the
// trigger waits for the damage to finish (AtBatchEnd, ThatMuchDamage), reads
// the total the creature was dealt at once, and the creature deals it. The
// controller is read as the damage lands, or as the creature last was if it
// died of it (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc1c2203-4cc6-4131-9d3c-6f3da88b97ce",
		Name:         "Spiteful Shadows",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
			},
			Key:        "Spiteful Shadows — enchanted creature deals that much damage to its controller",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.S82OldB440SpitefulShadowsReflect,
		}},
	})
}
