package game

import (
	"strings"

	"github.com/google/uuid"
)

// meld.go — CR 701.42 and CR 712.4: two cards that become one
// permanent. ADR 0145.
//
// A meld pair is two cards whose back faces are the two halves of one
// oversized card (CR 712.4). One of them prints an ability that exiles
// both and "melds them" (CR 701.42a): they come back onto the
// battlefield as a SINGLE OBJECT represented by two cards, with only
// the characteristics of the combined back face (CR 712.8g). When that
// object leaves the battlefield, one permanent leaves and two cards
// arrive in the new zone (CR 712.21).
//
// # The representation
//
// Card is one card, and nearly everything in the engine is keyed by a
// Card's InstanceID. Rather than teach every one of those readers that
// a permanent may be two cards, the melded permanent IS one Card — a
// value whose printed fields are the combined back face's, whose
// OracleID and ScryfallID are the back face's own Scryfall record (so
// the catalog entry, the image and the coverage flag all key on it
// with nothing special), and which carries the two cards it is made of
// in MeldedFrom. A copy effect never reads MeldedFrom (PrintedValues
// has no slot for it), so a copy of a melded permanent is the back
// face and nothing else, and its mana value is 0 (CR 712.8g, 202.3c).
//
// The printed back face comes from the deck import: every meld card's
// Scryfall record names its meld_result, and the importer stamps that
// record onto Card.Meld (MeldPrint). Both halves carry it, so either
// can be the one whose ability fires.
//
// # Leaving the battlefield
//
// MoveCard — the one function every zone change comes down to — splits
// a melded object into its cards whenever it lands anywhere but the
// battlefield or the stack (splitMeldedForMove). The first card of
// MeldedFrom, the CARRIER, takes the melded object's InstanceID: every
// caller that moved "the card with this ID" and then looks for it in
// the destination finds a real card there, and every event the move
// emits (one zone move, one leaves-the-battlefield) names one object,
// which is CR 712.21's "one permanent leaves the battlefield". The
// other card keeps the ID it had in exile and gets its own zone-move
// event from the two exits that can carry a melded permanent
// (landMeldPassengersLocked), so "whenever a card is put into a
// graveyard from anywhere" sees two cards while "whenever a creature
// dies" sees one death (CR 712.21's example).
//
// The carrier is the commander when either card is one, so CR 903.9a's
// question after a melded commander dies is asked of the right card
// with no special case; CR 903.9c's command-zone redirect sends the
// other card on to the zone the move was going to
// (sendMeldPassengersOnLocked).

// LayoutMeld is Scryfall's layout for a meld card and for the combined
// back face's own record.
const LayoutMeld = "meld"

// MeldPrint is the printed meld data of a meld card: the combined back
// face its pair forms (CR 712.4). Plain data, stamped at import.
type MeldPrint struct {
	// ResultOracleID is the combined back face's own oracle ID — the
	// catalog key of the melded permanent ("Urza, Planeswalker").
	ResultOracleID string `json:"resultOracleId"`

	// ResultScryfallID is the printing the back face's art is served
	// from.
	ResultScryfallID string `json:"resultScryfallId,omitempty"`

	// Result is the combined back face's printed characteristics, in
	// the engine's face shape: name, type line, colours, power and
	// toughness, loyalty, keywords. Its ManaCost is empty: a back face
	// prints none, and a melded permanent's mana value is read from
	// its cards' front faces instead (manaCostForValue).
	Result Face `json:"result"`

	// ResultNeedsEffect is Card.NeedsEffect for the back face: whether
	// what it prints needs a catalog entry to be automated. The
	// melded permanent takes it, so the "resolve by hand" badge is the
	// back face's answer, not the front face's.
	ResultNeedsEffect bool `json:"resultNeedsEffect,omitempty"`

	// ResultProducedMana is the back face's Scryfall produced_mana.
	ResultProducedMana []string `json:"resultProducedMana,omitempty"`
}

// clone deep-copies the print, nil for nil.
func (m *MeldPrint) clone() *MeldPrint {
	if m == nil {
		return nil
	}
	out := *m
	out.Result.Colors = copyStringSlice(m.Result.Colors)
	out.Result.Keywords = copyStringSlice(m.Result.Keywords)
	out.ResultProducedMana = copyStringSlice(m.ResultProducedMana)
	return &out
}

// IsMelded reports whether this object is a melded permanent: one
// object represented by two cards (CR 712.4a).
func (c Card) IsMelded() bool { return len(c.MeldedFrom) > 0 }

// IsMeldCard reports whether this card is half of a meld pair — a CARD
// whose printing is a meld card, not an object that is copying one.
func (c Card) IsMeldCard() bool { return c.Meld != nil && c.Meld.ResultOracleID != "" }

// CanMeld reports whether two cards form a meld pair (CR 701.42b): two
// meld cards, neither a token, whose printings name the same combined
// back face and that are not the same card. A Clone copying Urza, Lord
// Protector is not a meld card (Card.Meld is the CARD's), so it can be
// exiled by the ability but not melded.
func CanMeld(a, b Card) bool {
	if !a.IsMeldCard() || !b.IsMeldCard() || a.IsToken() || b.IsToken() || a.IsMelded() || b.IsMelded() {
		return false
	}
	if a.InstanceID == b.InstanceID || a.OracleID == b.OracleID {
		return false
	}
	return a.Meld.ResultOracleID == b.Meld.ResultOracleID
}

// meldedPermanentOf builds the melded object from its two cards, as
// they sit in exile (CR 701.42a, 712.14c). The carrier — `a`, or `b`
// when only `b` is a commander — lends the object its InstanceID, its
// owner and its controller; the printed fields are the combined back
// face's. Everything battlefield-side is left at its zero value, as it
// is on any card in exile: the entry pipeline stamps the rest.
func meldedPermanentOf(a, b Card) Card {
	if b.IsCommander && !a.IsCommander {
		a, b = b, a
	}
	a.CommanderReturnDue = false
	b.CommanderReturnDue = false
	// Neither card is a token (CanMeld), so neither carries
	// card-instance abilities; clearing them keeps MeldedFrom pure
	// carried data (closure_fields.txt classifies the route `rebuilt`).
	a.ManaAbilities, a.ActivatedAbilities = nil, nil
	b.ManaAbilities, b.ActivatedAbilities = nil, nil
	mp := a.Meld
	m := Card{
		InstanceID:        a.InstanceID,
		ObjectEpoch:       a.ObjectEpoch,
		Owner:             a.Owner,
		Controller:        a.Controller,
		OracleID:          mp.ResultOracleID,
		ScryfallID:        mp.ResultScryfallID,
		Name:              mp.Result.Name,
		TypeLine:          mp.Result.TypeLine,
		Colors:            copyStringSlice(mp.Result.Colors),
		Power:             mp.Result.Power,
		Toughness:         mp.Result.Toughness,
		VariableToughness: mp.Result.VariableToughness,
		StartingLoyalty:   mp.Result.StartingLoyalty,
		StartingDefense:   mp.Result.StartingDefense,
		Keywords:          copyStringSlice(mp.Result.Keywords),
		ProducedMana:      copyStringSlice(mp.ResultProducedMana),
		ColorIdentity:     unionColors(a.ColorIdentity, b.ColorIdentity),
		NeedsEffect:       mp.ResultNeedsEffect,
		Layout:            LayoutMeld,
		// CR 903.3b: if a commander is melded, the melded permanent is
		// that player's commander.
		IsCommander: a.IsCommander || b.IsCommander,
		KnownBy:     copyKnownBy(a.KnownBy),
		MeldedFrom:  []Card{a, b},
	}
	return m
}

// copyKnownBy copies a knowledge set, nil for none.
func copyKnownBy(in map[uuid.UUID]bool) map[uuid.UUID]bool {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// meldedManaCost is the cost a melded permanent's mana value is read
// from: its cards' front-face costs, concatenated (CR 712.8g, 202.3c).
func meldedManaCost(c Card) string {
	var b strings.Builder
	for _, part := range c.MeldedFrom {
		b.WriteString(part.ManaCost)
	}
	return b.String()
}

// splitMeldedForMove is MoveCard's half of CR 712.21: the cards a
// melded object becomes as it lands in `dst`, in the order they are
// pushed (the last one ends on top). The carrier takes the object's
// InstanceID and its new ObjectEpoch; the other card keeps its own ID
// and starts a new object too (CR 400.7). Each is a card in a zone
// that is not the battlefield, so the per-card clean state it was
// melded with is what it lands with, plus CR 903.9a's mark.
func splitMeldedForMove(m Card, dst ZoneKind) []Card {
	out := make([]Card, 0, len(m.MeldedFrom))
	for i, part := range m.MeldedFrom {
		part.MeldedFrom = nil
		part.MeldSplitFrom = ObjectRef{}
		if i == 0 {
			part.InstanceID = m.InstanceID
			part.ObjectEpoch = m.ObjectEpoch
		} else {
			part.ObjectEpoch++
		}
		part.Tapped = false
		part.Counters = nil
		part.effective = nil
		part.CommanderReturnDue = commanderReturnDueOn(part, dst)
		out = append(out, part)
	}
	// CR 712.21c: the carrier remembers the card it left with.
	if len(out) == 2 {
		out[0].MeldSplitFrom = ObjectRef{ID: out[1].InstanceID, Epoch: out[1].ObjectEpoch}
	}
	return out
}

// meldPassengers is the cards of a melded object other than its
// carrier, as MoveCard pushed them: the IDs a caller of MoveCard owes
// a landing (landMeldPassengersLocked). Empty for every other card.
func meldPassengers(moved Card) []uuid.UUID {
	if len(moved.MeldedFrom) < 2 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(moved.MeldedFrom)-1)
	for _, part := range moved.MeldedFrom[1:] {
		out = append(out, part.InstanceID)
	}
	return out
}

// landMeldPassengersLocked finishes a melded permanent's exit for the
// card that is not its carrier, once MoveCard has put both into `dst`
// (CR 712.21): who may see it in its new zone (CR 401.2 for a library),
// its place in a library when the move asked for the bottom or a depth
// (both cards go there — CR 712.21c), and its own zone-move event, so a
// "put into a graveyard from anywhere" ability sees two cards arrive.
// It gets no leaves-the-battlefield event: one permanent left, and the
// carrier's event is that one.
//
// askOrder puts CR 712.21a's question to the owner when the two landed
// together in their graveyard or library: the caller passes false for
// a library that is about to be shuffled, where no order survives.
//
// Caller must hold g.mu.
func (g *Game) landMeldPassengersLocked(moved Card, src ZoneKind, dst *Zone, actor uuid.UUID, toBottom bool, depth int, askOrder bool) {
	defer func() {
		if askOrder {
			g.queueMeldOrderLocked(moved, dst)
		}
	}()
	for _, id := range meldPassengers(moved) {
		if !dst.Contains(id) {
			continue
		}
		if dst.Kind == ZoneLibrary {
			for i := range dst.Cards {
				if dst.Cards[i].InstanceID == id {
					dst.Cards[i].ClearKnown()
					break
				}
			}
			if toBottom || depth > 1 {
				if c, err := dst.Remove(id); err == nil {
					if toBottom {
						dst.PushBottom(c)
					} else {
						dst.InsertFromTop(c, depth)
					}
				}
			}
		} else {
			g.markCardKnownInZoneLocked(dst, id)
		}
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   actor,
			CardID:  id,
			OldZone: src,
			NewZone: dst.Kind,
		})
	}
}

// sendMeldPassengersOnLocked is CR 903.9c: a melded commander whose
// owner put it into the command zone instead sends only the commander
// card there; each card of it that is not a commander goes to the zone
// the move was going to. MoveCard has put both cards into the command
// zone; this takes the others out again and on to `dst` (their owner's
// zone of that kind, exile for an owner who has left). A no-op for any
// move that was not a melded permanent's.
//
// Caller must hold g.mu.
func (g *Game) sendMeldPassengersOnLocked(moved Card, command *Zone, dst ZoneKind, actor uuid.UUID) {
	if command == nil || command.Kind != ZoneCommand {
		return
	}
	for _, id := range meldPassengers(moved) {
		card, ok := g.cardInZoneLocked(command, id)
		if !ok || card.IsCommander {
			continue
		}
		var to *Zone
		if p := g.playerByIDLocked(card.Owner); p != nil {
			switch dst {
			case ZoneHand:
				to = p.Hand
			case ZoneLibrary:
				to = p.Library
			case ZoneGraveyard:
				to = p.Graveyard
			}
		}
		if dst == ZoneExile || to == nil {
			to = g.Exile
		}
		if _, err := MoveCard(command, to, id); err != nil {
			continue
		}
		passenger := Card{MeldedFrom: []Card{{}, card}}
		g.landMeldPassengersLocked(passenger, ZoneBattlefield, to, actor, false, 0, false)
	}
}

// MeldForEffect is CR 701.42a's keyword action, the second half of
// every printed meld ability: "exile them, then meld them into <back
// face>". `sourceID` is the permanent whose ability it is and
// `partnerID` the other permanent the ability found; `controller` is
// who the melded permanent enters under (uuid.Nil: its owner).
//
// The condition every printed meld ability states first — "if you
// both own and control <this> and a <type> named <partner>" — is the
// card's to check (effects.Meld), because it is printed text, and the
// printed text differs (Mishra, Claimed by Gix also wants both
// attacking). This verb is the rest:
//
//  1. Both permanents are exiled, as one simultaneous exit
//     (ExileCardsThenForEffect), so a commander is offered the command
//     zone like any other exiled commander (CR 903.9a) and the rest
//     waits for the answer.
//  2. If both reached exile and form a meld pair (CanMeld), they are
//     put onto the battlefield melded (CR 712.14c): the melded object
//     replaces the carrier in exile for the width of one locked
//     mutation and returns through the ordinary exile entry, which
//     mints the new object (CR 400.7) and runs the CR 614 pipeline, so
//     its "when this enters" abilities fire and a replacement that
//     taps or redirects an entering permanent applies to it.
//  3. Otherwise they stay in exile (CR 701.42c): a token copy of Graf
//     Rats is exiled and ceases to exist, the Midnight Scavengers
//     exiled beside it stays there.
//
// An entry that is cancelled or redirected puts the two cards back in
// exile as two cards (unmeldInExileLocked), so a melded object is never
// left anywhere but the battlefield.
//
// Caller must hold g.mu.
func (g *Game) MeldForEffect(sourceID, partnerID, controller uuid.UUID) error {
	if sourceID == partnerID || findBattlefieldCard(g, sourceID) == nil || findBattlefieldCard(g, partnerID) == nil {
		return nil
	}
	return g.ExileCardsThenForEffect([]uuid.UUID{sourceID, partnerID}, func(g *Game, exiled []uuid.UUID) error {
		if len(exiled) != 2 || g.Exile == nil {
			return nil
		}
		a, okA := g.cardInZoneLocked(g.Exile, sourceID)
		b, okB := g.cardInZoneLocked(g.Exile, partnerID)
		if !okA || !okB || !CanMeld(a, b) {
			return nil
		}
		m := meldedPermanentOf(a, b)
		if err := g.meldInExileLocked(m); err != nil {
			return err
		}
		names := m.MeldedFrom[0].Name + " and " + m.MeldedFrom[1].Name
		return g.ReturnFromExileToBattlefieldThenForEffect(m.InstanceID, controller, false, func(g *Game, entered uuid.UUID) error {
			if entered == uuid.Nil {
				g.unmeldInExileLocked(m.InstanceID)
				return nil
			}
			c := findBattlefieldCard(g, entered)
			if c == nil {
				return nil
			}
			g.EmitEvent(Event{
				Kind:   EventMeld,
				CardID: entered,
				Actor:  c.Controller,
				Label:  names,
			})
			return nil
		})
	})
}

// meldInExileLocked swaps the two exiled cards for the melded object
// that is about to enter: the carrier's slot holds it, the other card
// is lifted out (it travels inside MeldedFrom now).
//
// Caller must hold g.mu.
func (g *Game) meldInExileLocked(m Card) error {
	other := m.MeldedFrom[1].InstanceID
	if _, err := g.Exile.Remove(other); err != nil {
		return err
	}
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == m.InstanceID {
			g.Exile.Cards[i] = m
			return nil
		}
	}
	return ErrCardNotFound
}

// unmeldInExileLocked undoes meldInExileLocked for an entry that did
// not happen: a melded object still in exile becomes its two cards
// again, each where a card exiled by the meld ability would be.
//
// Caller must hold g.mu.
func (g *Game) unmeldInExileLocked(id uuid.UUID) {
	if g.Exile == nil {
		return
	}
	m, ok := g.cardInZoneLocked(g.Exile, id)
	if !ok || !m.IsMelded() {
		return
	}
	if _, err := g.Exile.Remove(id); err != nil {
		return
	}
	for _, part := range m.MeldedFrom {
		if part.InstanceID == m.MeldedFrom[0].InstanceID {
			part.InstanceID = id
		}
		part.MeldedFrom = nil
		g.Exile.PushTop(part)
	}
}

// depthIf is depth when ok, else 0: a library depth the move asked for
// applies only when the move was not redirected away from it.
func depthIf(ok bool, depth int) int {
	if !ok {
		return 0
	}
	return depth
}

// meldSplitPartnerInExileLocked is CR 712.21c for the exile returns: the
// other card a melded permanent became as it was exiled, when `cardID`
// is the card that carried it and that other card is still the object
// the split put into exile. uuid.Nil otherwise.
//
// Caller must hold g.mu.
func (g *Game) meldSplitPartnerInExileLocked(cardID uuid.UUID) uuid.UUID {
	if g.Exile == nil {
		return uuid.Nil
	}
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok || c.MeldSplitFrom.ID == uuid.Nil {
		return uuid.Nil
	}
	p, ok := g.cardInZoneLocked(g.Exile, c.MeldSplitFrom.ID)
	if !ok || p.ObjectEpoch != c.MeldSplitFrom.Epoch {
		return uuid.Nil
	}
	return p.InstanceID
}

// objectRefPtr and objectRefValue carry Card.MeldSplitFrom through the
// snapshot as an omitted key when it is zero, which it is on almost
// every card.
func objectRefPtr(r ObjectRef) *ObjectRef {
	if r.ID == uuid.Nil {
		return nil
	}
	return &r
}

func objectRefValue(r *ObjectRef) ObjectRef {
	if r == nil {
		return ObjectRef{}
	}
	return *r
}

// meldOrderThen is CR 712.21a's continuation: "If a melded permanent is
// put into its owner's graveyard or library, that player may arrange
// the two cards in any order." Option 0 keeps the order the move left;
// option 1 swaps the two cards. Carry is [the card on top now, the
// other one].
var meldOrderThen = RegisterOptionPickThen("meld/order", meldOrderChosen)

func meldOrderChosen(g *Game, r OptionPicked) error {
	if r.Option == nil || r.Index != 1 || len(r.Carry) != 2 {
		return nil
	}
	z := g.findCardZoneLocked(r.Carry[0])
	if z == nil || (z.Kind != ZoneGraveyard && z.Kind != ZoneLibrary) || !z.Contains(r.Carry[1]) {
		return nil
	}
	i, j := -1, -1
	for k := range z.Cards {
		switch z.Cards[k].InstanceID {
		case r.Carry[0]:
			i = k
		case r.Carry[1]:
			j = k
		}
	}
	if i >= 0 && j >= 0 {
		z.Cards[i], z.Cards[j] = z.Cards[j], z.Cards[i]
	}
	return nil
}

// queueMeldOrderLocked asks a melded permanent's owner which of its two
// cards goes on top, when both have just landed in that player's
// graveyard or library (CR 712.21a). The question names the cards,
// which were public a moment ago; what a library hides is the order the
// owner picks, which the answer does not broadcast.
//
// Caller must hold g.mu.
func (g *Game) queueMeldOrderLocked(moved Card, dst *Zone) {
	if dst == nil || (dst.Kind != ZoneGraveyard && dst.Kind != ZoneLibrary) || len(moved.MeldedFrom) != 2 {
		return
	}
	carrier, passengers := moved.InstanceID, meldPassengers(moved)
	if len(passengers) != 1 {
		return
	}
	idx := map[uuid.UUID]int{}
	names := map[uuid.UUID]string{}
	for k, c := range dst.Cards {
		if c.InstanceID == carrier || c.InstanceID == passengers[0] {
			idx[c.InstanceID] = k
			names[c.InstanceID] = c.Name
		}
	}
	if len(idx) != 2 || dst.Owner == uuid.Nil {
		return
	}
	top, lower := carrier, passengers[0]
	if idx[lower] > idx[top] {
		top, lower = lower, top
	}
	where := "graveyard"
	if dst.Kind == ZoneLibrary {
		where = "library"
	}
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  dst.Owner,
		Source:   carrier,
		Question: "Arrange the two cards of " + moved.Name + " in your " + where + ": which is above the other?",
		Options: []ChoiceOption{
			{Label: names[top] + " above " + names[lower]},
			{Label: names[lower] + " above " + names[top]},
		},
		ThenKey: meldOrderThen,
		Carry:   []uuid.UUID{top, lower},
	})
}
