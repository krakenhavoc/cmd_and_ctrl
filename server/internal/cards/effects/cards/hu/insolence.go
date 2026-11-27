package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Insolence — Enchantment — Aura {2}{R} (Planeshift, #5615):
//
//	"Enchant creature
//	 Whenever enchanted creature becomes tapped, this Aura deals 2 damage to
//	 that creature's controller."
//
// Brink of Disaster's "becomes tapped" watcher (a tap effect, a tap cost or an
// attack all count) with 2 damage from the Aura to the creature's controller,
// who may not be the Aura's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33a6d78a-a3c7-44f1-8ce9-d6632123fec1",
		Name:         "Insolence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Insolence — 2 damage to that creature's controller",
			Effect:    effects.S82OldB636InsolenceDamage,
		}},
	})
}
