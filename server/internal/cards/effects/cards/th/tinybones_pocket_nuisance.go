package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tinybones, Pocket Nuisance — Legendary Creature — Skeleton Rogue
// {2}{B}, 2/1:
//
//	"When Tinybones enters, each opponent discards a card.
//	 Whenever a player discards one or more cards, Tinybones deals 1
//	 damage to each opponent."
//
// The engine emits one discard event per card, so the second trigger is
// a once-per-batch one, keyed per discarding player (CR 603.2c): one
// player discarding two cards at once is one trigger, two players each
// discarding is two. Its own ETB discard counts, so the opponents each
// take 1 as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97dcf9fb-3f2f-4aa6-959a-e3c4889c1673",
		Name:         "Tinybones, Pocket Nuisance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tinybones, Pocket Nuisance — each opponent discards a card", effects.EachOpponentDiscardsOne),
			effects.RfCreatureEOncePerDiscarder(effects.On(game.EventDiscardCard, effects.AnyPlayer,
				"Tinybones, Pocket Nuisance — 1 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 1}))),
		},
	})
}
