package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warped Devotion — Enchantment {2}{B} (Eighth Edition, #5607):
//
//	"Whenever a permanent is returned to a player's hand, that player
//	 discards a card."
//
// One trigger per permanent that moves from the battlefield to a hand (CR
// 400.1), however it got there; the player whose hand it is chooses the
// discard as the trigger resolves.
//
// No simplification.
func init() {
	const label = "Warped Devotion — that player discards a card"
	effects.Register(effects.Spec{
		OracleID:     "22c49415-e8d2-4b4c-b021-57ee54670f86",
		Name:         "Warped Devotion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.S82OldB615PermanentReturnedToAHand,
			Key:       label,
			Build:     effects.S82OldB615BuildNamingTheHandOwner(label),
			Effect:    effects.S82OldB615ThatPlayerDiscards,
		}},
	})
}
