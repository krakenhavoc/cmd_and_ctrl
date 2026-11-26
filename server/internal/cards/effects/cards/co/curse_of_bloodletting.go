package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of Bloodletting — Enchantment — Aura Curse {3}{R}{R} (Dark Ascension,
// S82 OLD, #5571):
//
//	"Enchant player
//	 If a source would deal damage to enchanted player, it deals double that
//	 damage to that player instead."
//
// Fiendish Duo's doubling replacement (CR 614, CR 615.1), narrowed to the
// player this Curse is attached to as the damage would be dealt: every source,
// combat or not, and the Curse's own controller's included. Damage to anything
// else, or to the player once the Curse has left, is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3383a72f-fb83-4b9a-aca6-b736ed929c2e",
		Name:         "Curse of Bloodletting",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 &&
					src.AttachedTo.Kind == game.TargetPlayer && src.AttachedTo.ID == ev.DamageTarget
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Curse of Bloodletting — double damage to enchanted player",
		}},
	})
}
