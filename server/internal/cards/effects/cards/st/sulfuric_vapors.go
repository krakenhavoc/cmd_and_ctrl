package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sulfuric Vapors — Enchantment {3}{R} (Urza's Saga, S82 OLD b723):
//
//	If a red spell would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.
//
// A replacement on damage (Jaya's +1, CR 614.1a, 120.3) whose source is a spell that is red as it would deal the damage: any player's, instants and sorceries; a red permanent's damage is not a spell's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40b14aaa-f29e-4e84-9bdc-bc2ac9762ce9",
		Name:         "Sulfuric Vapors",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.S82OldB723RedSpellDealsDamage,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Sulfuric Vapors: +1 damage from a red spell",
		}},
	})
}
