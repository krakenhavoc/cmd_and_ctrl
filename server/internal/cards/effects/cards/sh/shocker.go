package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shocker — Creature — Insect {1}{R}, 1/1 (Tempest, S82 OLD batch 766, #5628):
//
//	"Whenever this creature deals damage to a player, that player discards all
//	 the cards in their hand, then draws that many cards."
//
// Combat or noncombat damage to any player, its own controller included. The
// damaged player is stamped on the trigger as it goes on the stack, discards
// their whole hand, then draws as many cards as were really discarded.
//
// No simplification.
func init() {
	const label = "Shocker — that player discards all the cards in their hand, then draws that many cards"
	effects.Register(effects.Spec{
		OracleID:     "8981d1d1-e102-4a3d-96ad-3904c83d1611",
		Name:         "Shocker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtDamageToAPlayer,
			Key:       label,
			Build:     effects.RecordTheEventsTargetAsThePlayer(label),
			Effect:    effects.ThatPlayerDiscardsTheirHandThenDrawsThatMany("shocker/discarded"),
		}},
	})
}
