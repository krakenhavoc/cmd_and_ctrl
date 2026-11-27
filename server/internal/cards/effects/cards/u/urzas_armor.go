package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urza's Armor — Artifact {6} (Eighth Edition, S82 OLD batch 611, #5607):
//
//	"If a source would deal damage to you, prevent 1 of that damage."
//
// A CR 615 prevention effect on damage to the controller alone (Orbs of
// Warding's replacement without the creature restriction): any source, its
// own controller's included, each damage event losing 1, and damage of 1
// prevented entirely (CR 615.1a). Damage to the controller's permanents is
// untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5fa232a-c8f2-4532-88bf-60d1223181b0",
		Name:         "Urza's Armor",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 && ev.DamageTarget == src.Controller
			},
			Replace:    effects.S82OldB317PreventOneOfThatDamage,
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			Label:      "Urza's Armor — prevent 1 of that damage",
			RowLabel:   "If a source would deal damage to you, prevent 1 of that damage.",
		}},
	})
}
