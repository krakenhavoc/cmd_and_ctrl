package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b2_helpers.go — shared pieces of the Secrets of Strixhaven
// Commander batch 2 (#4834).

// equippedCreatureAttacksSearchBasicLand is "Whenever equipped creature
// attacks, you may search your library for a basic land card, put it
// onto the battlefield tapped, then shuffle." Sword of the Animist's
// trigger, shared with Bitterthorn, Nissa's Animus. `name` labels the
// trigger and the search, so the keys Sword of the Animist already had
// are unchanged.
//
// The condition is attachedCreatureAttacked: EventAttack carries the
// attacking creature and fires once per creature on its declaration, so
// one Equipment on one creature is one trigger per combat. The search is
// Optional, so the searcher may decline the land and the shuffle.
func equippedCreatureAttacksSearchBasicLand(name string) game.TriggeredAbility {
	return On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return attachedCreatureAttacked(ev, source)
	}, name+" — search for a basic land", func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:        item.Controller,
			Predicate:     IsBasicLand,
			Dest:          game.ZoneBattlefield,
			Limit:         1,
			TappedOnEntry: true,
			Optional:      true,
			Shuffle:       true,
			Reason:        name + " — search for a basic land card",
		}.Apply(NewContext(g, item))
	})
}

// replicateMana is "Replicate [mana]" (Changing Loyalty's {2}, CR
// 702.56a): the optional cost keyed game.ReplicateKey, mana only, paid
// any number of times. `max` is the per-announcement cap the energy and
// tapping forms carry (replicate.go); Register pairs it with Replicate().
func replicateMana(mana string, max int) game.AdditionalCost {
	if max < 1 {
		max = 1
	}
	return game.AdditionalCost{
		Optional: true,
		Key:      game.ReplicateKey,
		ManaCost: mana,
		Repeat:   max,
		Label:    "Replicate " + mana,
	}
}

// tokensYouControlExcept lists the tokens `controller` controls on the
// battlefield other than `except`, in battlefield order. Brudiclad's
// "each other token you control". Caller holds g.mu.
func tokensYouControlExcept(g *game.Game, controller, except uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.InstanceID != except && IsToken(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
