package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cauldron Familiar — Creature — Cat {B}, 1/1 (EDHREC rank 2974):
//
//	"When this creature enters, each opponent loses 1 life and you
//	 gain 1 life.
//	 Sacrifice a Food: Return this card from your graveyard to the
//	 battlefield."
//
// The Cat half of Cat-Oven. The enters trigger drains every opponent
// for 1 and gains the controller 1 — one life, not one per opponent,
// as printed (eachOpponentLosesLife then GainLife).
//
// DECLARED SIMPLIFICATION, weaker than printed: the second ability
// is not offered. It is an activated ability that works from the
// GRAVEYARD, and the engine's CR 602 abilities are battlefield-only
// (ActivateCatalogAbility looks the source up on the battlefield);
// there is no zone slot on an activated ability the way CastableZones
// gives a spell one. An ability with no way to activate is left out
// rather than faked, and the card is still recognisably itself: a
// one-mana Cat that drains on entry, and comes back through any
// reanimation the deck already has. The loop lands when activated
// abilities gain a from-graveyard zone.
func init() {
	Register(Spec{
		OracleID:     "a7dc2e62-1c50-4ed7-b71f-2d782a447a5e",
		Name:         "Cauldron Familiar",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Cat can't be returned from your graveyard by sacrificing a Food — abilities can only be activated from the battlefield."},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cauldron Familiar — each opponent loses 1 life and you gain 1 life", drainEachOpponent),
		},
	})
}
