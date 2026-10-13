package game

import "github.com/google/uuid"

// exiled_with.go — CR 607.2a's "exiled with [this permanent]" for a
// card a REPLACEMENT EFFECT exiled (#2530, ADR 0066's 2026-10-07
// amendment), and since ADR 0155 PR 3 (#3201) for a card a RESOLVING
// ABILITY exiled through ExileLinkedForEffect.
//
// Valgavoth, Terror Eater is the first card: "if a card you didn't
// control would be put into an opponent's graveyard from anywhere,
// exile it instead. During your turn, you may play cards exiled with
// Valgavoth." The exile is a replacement, which runs before the card
// moves, so there is no moment inside it at which a card in exile
// exists to be marked. The link therefore travels on the event
// (ReplacementEvent.ExiledWith) and lands with the card, here.
//
// It is a card-carried link rather than the effects package's event-log
// record (b27ExiledWith) because that record is keyed on an ABILITY
// resolving, and a replacement is not one: nothing is resolving, there
// is no EventResolve to open a window on. It is also read by a pure-data
// PermissionFilter, which has a Card in hand and no log to walk.
//
// ADR 0155 §8 extends the stamp to a resolving ability's exile. A
// "has all activated abilities of all cards exiled with it" borrow reads
// its lenders on every layer pass, and walking the event log there is
// far too slow, so the link has to be on the card. ExileLinkedForEffect
// is the verb: it is the ordinary exile route with the source object on
// it (zoneRoute.ExiledWith), which seeds ReplacementEvent.ExiledWith, so
// the stamp lands through the same two finishers as Valgavoth's and only
// when the card really reaches exile.

// ExileLinkedForEffect exiles `cardID` from wherever it is, linked to
// the source OBJECT `source` (CR 607.2a): once it has landed in exile
// its Card.ExiledWith names {source.ID, source.Epoch}. "Exile target
// creature" on Dark Impostor, "exile target artifact card from a
// graveyard" on Myr Welder. `source` is the resolving item's
// SourceObject (Game.SourceObjectForEffect), so a source that left and
// came back is a new object with no claim on the card (CR 400.7).
//
// The FIRE-AND-FORGET form, ExileCardForEffect's contract: nil means no
// error, never "it is in exile". A caller that reads the outcome uses
// ExileLinkedThenForEffect. A zero source exiles with no link.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileLinkedForEffect(source ObjectRef, cardID uuid.UUID) error {
	r := exileRoute
	r.CardID = cardID
	r.ExiledWith = PermissionCardRef(source)
	_, err := g.routeCardToZoneLocked(r)
	return err
}

// ExileLinkedThenForEffect is ExileLinkedForEffect over several cards as
// one simultaneous exile, handing `then` the ones that reached exile
// (ExileCardsThenForEffect's contract): "exile up to three target land
// cards" on Steward of the Harvest, and any linked exile with a "then".
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileLinkedThenForEffect(source ObjectRef, ids []uuid.UUID, then func(g *Game, exiled []uuid.UUID) error) error {
	r := exileRoute
	r.ExiledWith = PermissionCardRef(source)
	return g.routeAllThenLocked(r, ids, then)
}

// ExiledWithForEffect lists the cards in exile linked to the source
// OBJECT `source` (Card.ExiledWith == source), in exile order, so the
// last is the one exiled most recently. Pointers into the exile zone, for
// a reader inside a layer pass (a borrow's Lenders); a caller that keeps
// them past the lock copies them.
//
// Caller must hold g.mu (read or write).
func (g *Game) ExiledWithForEffect(source PermissionCardRef) []*Card {
	if g.Exile == nil || source.ID == uuid.Nil {
		return nil
	}
	var out []*Card
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].ExiledWith == source {
			out = append(out, &g.Exile.Cards[i])
		}
	}
	return out
}

// stampExileCounterLocked puts one `kind` counter on the card that has
// just landed in `dst`, when `dst` is exile and the move declared one
// (ReplacementEvent.ExileCounter). A no-op for every other destination,
// for an empty kind, and for a card that is not in `dst`.
func stampExileCounterLocked(dst *Zone, cardID uuid.UUID, kind string) {
	if dst == nil || dst.Kind != ZoneExile || kind == "" {
		return
	}
	for i := range dst.Cards {
		if dst.Cards[i].InstanceID == cardID {
			if dst.Cards[i].Counters == nil {
				dst.Cards[i].Counters = map[string]int{}
			}
			dst.Cards[i].Counters[kind]++
			return
		}
	}
}

// stampExiledWithLocked writes the "exiled with" link onto the card
// that has just landed in `dst`, when `dst` is exile and the move
// declared a link. A no-op for every other destination, for the zero
// ref, and for a card that is not in `dst` (a failed move).
//
// Called after MoveCard, which has already cleared any older link:
// the stamp names the move that just happened and no earlier one.
func stampExiledWithLocked(dst *Zone, cardID uuid.UUID, ref PermissionCardRef) {
	if dst == nil || dst.Kind != ZoneExile || ref.ID == uuid.Nil {
		return
	}
	for i := range dst.Cards {
		if dst.Cards[i].InstanceID == cardID {
			dst.Cards[i].ExiledWith = ref
			return
		}
	}
}
