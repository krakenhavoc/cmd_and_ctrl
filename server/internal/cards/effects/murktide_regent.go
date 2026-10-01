package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murktide Regent — Creature — Dragon 3/3 {5}{U}{U}:
//
//	"Delve
//	 Flying
//	 This creature enters with a +1/+1 counter on it for each instant
//	 and sorcery card exiled with it.
//	 Whenever an instant or sorcery card leaves your graveyard, put a
//	 +1/+1 counter on this creature."
//
// ADR 0100 sub-PR 2's first card. "Exiled with it" is CR 607.2q's link
// between the delve ability printed on the spell and the permanent it
// becomes: only the cards exiled to pay for THIS spell, and only while
// they are still in exile as those objects (CR 400.7). The entry clause
// is CountersPerDelved over game.CastCounts.Delved, so the counters are
// part of the entry (CR 614.1c): Doubling Season and Hardened Scales
// see them, and so does anything that watches counters being put on a
// permanent.
//
// The second ability is one trigger per card, not per batch: a Treasure
// Cruise that delves five instants puts five counters on it. Its own
// delve does not trigger it, because it is a spell on the stack while
// those cards leave (ADR 0100 §1). A card cast from the graveyard
// (flashback, escape) leaves it too, and its cast is the event
// (b16CardLeftYourGraveyard). "Your graveyard" is the graveyard you own.
//
// A CR 707.10 copy of the spell pays nothing, so the token it becomes
// enters with no counters from this clause (PaidCost.Delved is not
// copied, ADR 0100 sub-PR 1).
//
// No simplification.

const murktideRegentOracle = "7f993ac7-c2cd-413c-a106-2c051a77ebf6"

func init() {
	Register(Spec{
		OracleID:        murktideRegentOracle,
		Name:            "Murktide Regent",
		Completeness:    CompletenessFull,
		Delve:           true,
		PrintedKeywords: []string{"flying"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersPerDelved(game.CounterPlusOne, b08IsInstantOrSorceryCard),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventZoneMove, game.EventCast},
				murktideInstantOrSorceryLeftYourGraveyard,
				"Murktide Regent — put a +1/+1 counter on this creature",
				putCounterOnSelf),
		},
	})
}

// murktideInstantOrSorceryLeftYourGraveyard is "whenever an instant or
// sorcery card leaves your graveyard".
func murktideInstantOrSorceryLeftYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b16CardLeftYourGraveyard(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && wasInstantOrSorceryCardInAGraveyard(c)
}

// wasInstantOrSorceryCardInAGraveyard asks the type question of a card
// that has just LEFT a graveyard, as the card it was there (CR 603.10a:
// a leaves-the-graveyard trigger looks back in time).
//
// The card is looked up where it went, and for a single-faced card
// that is the same answer. A multi-faced one can have changed: a
// modal double-faced card cast from a graveyard as its back face, or
// an adventurer card cast as its Adventure, is on the stack as the
// half that was cast. In a graveyard it had only its front face's
// characteristics (CR 712.8a, 715.4), or, for a split card, both
// halves' together (CR 709.4), so those are what is asked here — never
// the half it left as, which would count an adventurer creature card
// as an instant.
func wasInstantOrSorceryCardInAGraveyard(c game.Card) bool {
	if !c.IsMultiFace() {
		return b08IsInstantOrSorceryCard(c)
	}
	faces := c.Faces[:1]
	if game.IsSplitCard(c) {
		faces = c.Faces
	}
	for _, f := range faces {
		if b08IsInstantOrSorceryCard(game.Card{TypeLine: f.TypeLine}) {
			return true
		}
	}
	return false
}
