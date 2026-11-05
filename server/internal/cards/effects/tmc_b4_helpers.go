package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b4_helpers.go — the shared pieces of the TMNT Commander batch 4
// cards (#5011): squad (CR 702.157).
//
//	"Squad [cost]" means "As an additional cost to cast this spell, you
//	 may pay [cost] any number of times" and "When this creature enters,
//	 if its squad cost was paid, create a token that's a copy of it for
//	 each time its squad cost was paid." (CR 702.157a)
//
// Two halves, each an existing seam, the replicate shape (replicate.go):
//
//   - The cost is an optional cost (ADR 0073) keyed SquadKey and paid any
//     number of times, the multikicker shape. The cast path sums it and
//     the resolution carries the paid record onto the permanent
//     (Card.Provenance.OptionalCosts, CR 400.7d), so a reanimated or
//     blinked creature, and every token copy (which was never cast),
//     paid nothing and triggers nothing.
//   - The trigger watches the creature's own entry. Its intervening if
//     (CR 603.4) is the paid count, which cannot change once the
//     creature is on the battlefield; the count is read as the trigger is
//     built and carried on the item's XValue, so a squad trigger waiting
//     on the stack is a restore point, and a creature that left in
//     response is still copied from last-known information (CR 608.2h,
//     the ruling). The copies are one creation (CR 608.2f), so they enter
//     at the same time (the ruling).
//
// A card declares both:
//
//	OptionalCosts: []game.AdditionalCost{SquadMana("{3}", 20)},
//	Triggered:     []game.TriggeredAbility{Squad("<name>")},

// SquadKey is the optional-cost key of squad. An on-disk identity (it
// names the paid entry on a restored card): never renamed, never reused.
const SquadKey = "squad"

// KeywordSquad is the printed keyword the squad trigger row carries.
const KeywordSquad = "squad"

// SquadMana is "Squad [mana]": the optional cost, paid any number of
// times. `max` is the engine's cap on one announcement, multikicker's
// reason (Multikicker): squad is unbounded in paper, an announcement is
// not. Say so in the card's comment.
func SquadMana(mana string, max int) game.AdditionalCost {
	if max < 1 {
		max = 1
	}
	return game.AdditionalCost{
		Optional: true,
		Key:      SquadKey,
		ManaCost: mana,
		Repeat:   max,
		Label:    "Squad " + mana,
	}
}

// squadPayments is how many times the permanent's squad cost was paid,
// read off the record its resolution carried onto it.
func squadPayments(c *game.Card) int {
	if c == nil {
		return 0
	}
	return game.OptionalCostTimesPaid(*c, c.Provenance.OptionalCosts, SquadKey)
}

// Squad is the keyword's trigger: "When this creature enters, if its
// squad cost was paid, create a token that's a copy of it for each time
// its squad cost was paid."
func Squad(name string) game.TriggeredAbility {
	label := name + " — squad"
	return game.TriggeredAbility{
		Keyword: KeywordSquad,
		Key:     label,
		Watches: []game.EventKind{game.EventETB},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && squadPayments(source) > 0
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			// A fill-in Build (ADR 0041 P9), replicate's: the count is a
			// fact of the cast; the effect is the row's.
			return &game.StackItem{
				Kind:         game.StackItemTriggered,
				Controller:   source.Controller,
				Owner:        source.Controller,
				SourceCardID: source.InstanceID,
				Label:        label,
				XValue:       squadPayments(source),
			}
		},
		Effect: squadEffect,
	}
}

// squadEffect creates the copies, one per payment, in one creation.
func squadEffect(g *game.Game, item *game.StackItem) error {
	if item.XValue <= 0 {
		return nil
	}
	return CreateTokenCopiesTogether{
		Controller: item.Controller,
		Copies:     []uuid.UUID{item.SourceCardID},
		N:          item.XValue,
	}.Apply(NewContext(g, item))
}
