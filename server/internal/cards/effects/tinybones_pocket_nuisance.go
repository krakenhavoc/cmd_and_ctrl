package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "97dcf9fb-3f2f-4aa6-959a-e3c4889c1673",
		Name:         "Tinybones, Pocket Nuisance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tinybones, Pocket Nuisance — each opponent discards a card", eachOpponentDiscardsOne),
			rfCreatureEOncePerDiscarder(On(game.EventDiscardCard, AnyPlayer,
				"Tinybones, Pocket Nuisance — 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error { return damageToEachOpponent(g, item, 1) })),
		},
	})
}
