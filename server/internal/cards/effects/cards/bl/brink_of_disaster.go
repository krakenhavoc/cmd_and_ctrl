package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brink of Disaster — Enchantment — Aura {2}{B}{B} (Magic 2012, S82 older sets
// batch 451, #5573):
//
//	"Enchant creature or land
//	 When enchanted permanent becomes tapped, destroy it."
//
// Chronic Flooding's "becomes tapped" watcher (a tap effect, a tap cost or an
// attack all count) with the permanent destroyed. The Aura's legality is the
// "creature or land" clause rerun every time state-based actions look.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae150b53-7c4e-48ce-abcf-cfe1e68bb4d5",
		Name:         "Brink of Disaster",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Or(effects.Creature(), effects.Land())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Brink of Disaster — destroy the enchanted permanent",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
					return nil
				}
				return effects.DestroyTarget{Target: item.Trigger.Event.CardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
