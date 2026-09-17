package game

import (
	"github.com/google/uuid"
)

// random_bottom.go — "put the rest on the bottom of your library in a
// random order" (#745; shared with effect RNG, #744).
//
// The clause is printed on cascade's reminder text and on most of the
// library-to-battlefield cards (The Regalia, Ureni, Atla Palani,
// Animist's Awakening). Until #745 it existed only as cascade's
// private bottomInRandomOrderLocked, which took cards from exile alone
// and moved them with a hand-rolled Remove/PushBottom that never
// opened the CR 614 window. This is the public version, and cascade
// now calls it.

// PutOnBottomInRandomOrderForEffect puts every card in `ids` that is
// still in a zone of kind `from` on the bottom of its owner's library
// in a random order.
//
// # Only from where the effect left them
//
// `from` is where the effect put or found the cards: ZoneExile for
// cascade's pile, ZoneLibrary for "the rest" of a reveal or a look.
// A card no longer in a zone of that kind is skipped. The IDs are
// usually saved before a prompt (cascade's may-cast, Lurking
// Predators' yes/no), and a card that has since been cast, drawn or
// moved by something else is a new object (CR 400.7) the clause no
// longer refers to — without the check, a stale ID would pull it back
// out of a hand, a graveyard or off the battlefield.
//
// # Randomness
//
// The order is one Fisher–Yates shuffle of `ids` as given (a repeated
// ID dropped), one draw on the actor's "random_order" stream of the
// game's keyed RNG (rng.go, ADR 0054 Decision 8). An undone effect
// bottoms the same cards in the same order when it is redone, a
// restored game continues the stream, and a draw on any other stream
// in between (a shuffle, a random discard) does not change it. There
// is no fallback source: a game that never called Start mints its key
// on the first draw. The shuffle runs before any card is looked up, so
// how much it draws from the stream depends only on the list and never
// on where the cards are.
//
// # How each card moves
//
//   - From any zone other than its owner's library (exile for
//     cascade): through routeCardToZoneLocked with ToBottom, the
//     engine's one exit path. That is what makes it a real zone
//     change — CR 614 replacements apply (a commander's owner is
//     offered the command zone, CR 903.9, and a declined offer still
//     lands on the bottom), the card becomes a new object with no
//     memory of exile (CR 400.7: its impulse grant and counters are
//     gone), and it arrives in a hidden zone with nobody knowing it.
//   - From its owner's library already (the revealed cards The Regalia
//     passed over never left the library): a REORDER, not a zone
//     change. routeCardToZoneLocked deliberately does nothing for a
//     move into the zone a card is already in, so this is handled
//     here: the card is lifted out and pushed on the bottom, no
//     replacement window opens and no zone-move event fires, because
//     nothing changed zones. Its knower set is cleared — a player who
//     watched it be revealed knows it is somewhere in the bottom N,
//     not where.
//
// A repeated ID is moved once. A commander's CR 903.9 prompt pauses
// only that card; the rest still go to the bottom, and the paused one
// lands on the bottom when its owner declines. An error on one card
// does not stop the rest; the first is returned.
//
// `actor` is the player the effect belongs to, stamped on the zone
// move events.
//
// Caller must hold g.mu.
func (g *Game) PutOnBottomInRandomOrderForEffect(actor uuid.UUID, from ZoneKind, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	order := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			order = append(order, id)
		}
	}
	swap := func(i, j int) { order[i], order[j] = order[j], order[i] }
	g.randForLocked(rngStream{kind: rngStreamRandomOrder, player: actor}).Shuffle(len(order), swap)
	var firstErr error
	for _, id := range order {
		src := g.findCardZoneLocked(id)
		if src == nil || src.Kind != from {
			continue
		}
		card, ok := g.cardInZoneLocked(src, id)
		if !ok {
			continue
		}
		if src.Kind == ZoneLibrary && src.Owner == card.Owner {
			c, err := src.Remove(id)
			if err != nil {
				continue
			}
			c.ClearKnown()
			src.PushBottom(c)
			continue
		}
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:   id,
			Dst:      ZoneLibrary,
			Actor:    actor,
			ToBottom: true,
		}); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// LookAtTopOfLibraryForEffect is "look at the top N cards of your
// library" as the first half of a card's text (Ureni, Armored
// Skyhunter, Explorer's Scope): it marks `playerID` — and only that
// player — a knower of each card and returns their IDs, top card
// first. Nothing moves and no prompt opens; what the card does with
// the cards is the caller's business.
//
// "Look at", not "reveal" (CR 701.20): the other seats learn that a
// look happened and how many cards it covered from the resolving
// ability, never which cards. The wire filter
// (protocol.FilterViewFor) keys off exactly the knower set this
// writes, and a choose_cards prompt over these cards drops its options
// for every seat but the chooser.
//
// A short library shows what it has and an empty one returns nil;
// looking is not drawing, so neither sets up the CR 704.5b loss.
//
// Caller must hold g.mu.
func (g *Game) LookAtTopOfLibraryForEffect(playerID uuid.UUID, n int) []uuid.UUID {
	if n <= 0 {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Library == nil {
		return nil
	}
	size := p.Library.Size()
	if n > size {
		n = size
	}
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		idx := size - 1 - i
		p.Library.Cards[idx].AddKnower(playerID)
		ids = append(ids, p.Library.Cards[idx].InstanceID)
	}
	return ids
}

// PutIntoGraveyardForEffect puts a card that is not on the battlefield
// into its owner's graveyard — Genesis Wave's "put all cards revealed
// this way that weren't put onto the battlefield into your graveyard".
//
// It is not a mill (CR 701.17a mills from the top of a library by
// count, and no mill payoff should see a Genesis Wave) and it is not a
// discard, so it emits a plain zone move. It goes through the shared
// exit path, so a commander's owner is offered the command zone
// (CR 903.9).
//
// A battlefield permanent is refused with ErrInvalidParam: leaving the
// battlefield for a graveyard is a destroy, a sacrifice or a
// state-based action, each with its own path and its own event, and
// none of them is this.
//
// Caller must hold g.mu.
func (g *Game) PutIntoGraveyardForEffect(cardID uuid.UUID) error {
	src := g.findCardZoneLocked(cardID)
	if src == nil {
		return ErrCardNotFound
	}
	if src.Kind == ZoneBattlefield {
		return ErrInvalidParam
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{CardID: cardID, Dst: ZoneGraveyard})
	return err
}
