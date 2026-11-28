package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Betrayal — Enchantment — Aura {U} (Visions, S82 older sets batch 790,
// #5631):
//
//	"Enchant creature an opponent controls
//	 Whenever enchanted creature becomes tapped, you draw a card."
//
// Insolence's "becomes tapped" watcher (a tap effect, a tap cost or an attack
// all count), the Aura's controller drawing. The enchant clause names an
// opponent's creature, and the Aura falls off if its host stops being one
// (CR 303.4d, 704.5m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c3d2e1f-a5da-47f4-aea8-5861ef7e1b90",
		Name:         "Betrayal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("enchant creature an opponent controls", effects.OpponentControls()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Betrayal — you draw a card",
			Effect:    effects.Old790YouDrawACard,
		}},
	})
}
