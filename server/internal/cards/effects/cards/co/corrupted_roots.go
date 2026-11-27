package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrupted Roots — Enchantment — Aura {B} (Conflux):
//
//	"Enchant Forest or Plains
//	 Whenever enchanted land becomes tapped, its controller loses 2 life."
//
// Contaminated Ground's trigger on a different enchant clause: "Forest or
// Plains" is read through the land's effective subtypes (CR 305.6), so a land
// that stops being one falls off as a state-based action (CR 704.5m). The
// trigger watches every way the land becomes tapped, for mana or an attack
// or any other reason (CR 701.26a), and its controller loses the life as the
// trigger resolves; the loss is not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c6f343a-0088-477a-a6b5-256783ad0c27",
		Name:         "Corrupted Roots",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.Or(effects.HasSubtype("Forest"), effects.HasSubtype("Plains"))),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Corrupted Roots — its controller loses 2 life",
			Effect:    effects.EnchantedPermanentsControllerLosesLife(2),
		}},
	})
}
