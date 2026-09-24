package protocol

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_frame.go projects game.EventRevealCards into GameView.Reveals
// — the broadcast counterpart to the controller-only look-at-cards
// prompt that scry, surveil and look_at_top ride.
//
// The two frames answer different questions and are deliberately
// shaped nothing alike.
//
// A LOOK AT is a PendingChoiceView: it names one chooser, it ships
// CardViews complete with instance IDs, and filterPendingChoices
// DROPS every option a non-chooser is not a knower of. It has to,
// because those options are cards in a hidden zone and a stable UUID
// is a correlation handle — the argument S31 sub-PR 0 settled for the
// public log and #513 had to settle again for the prompt.
//
// A REVEAL is this: it names no chooser, it goes to every seat
// identically, it is never redacted — and it carries NO INSTANCE ID
// AT ALL. Not for the revealed cards, and not for the card that
// revealed them.
//
// That last part is the whole design. A reveal makes a card public
// for that moment and to that extent. It does not make the library's
// order public, and it must not hand out a handle that survives the
// moment: Dark Confidant reveals the top card and puts it into hand,
// Fact or Fiction reveals five and buries some of them, and a tutor
// reveals a card that goes back onto a library about to be shuffled.
// Shipping the UUID would let any client — or any bot policy, which
// sees the same bytes under ADR 0033 §3 — write it down and recognise
// the card three turns later, in a zone it was never entitled to read.
// So the wire carries printed identity and nothing else: the name,
// the mana cost, the type line, and the Scryfall printing id the
// client needs to draw the face. Exactly what a player sitting at the
// table saw, and exactly nothing more.
//
// Because there is no handle and no hidden payload, there is nothing
// for FilterViewFor to redact, and it passes the field through
// untouched. That is not an omission — a reveal that one seat could
// not see would not be a reveal.
//
// The KNOWLEDGE half of a reveal is not here at all. It is the
// KnownBy set that game.RevealForEffect writes, it is permanent until
// a shuffle clears it, and it is what makes a revealed card stay
// readable in its owner's hand afterwards. This frame is the
// announcement, and it is bounded: see PublicRevealMax.

// PublicRevealMax bounds how many reveals ride one frame.
//
// A reveal is a table announcement, not a history — the public log is
// the history — so the window only has to be deep enough that a
// client which dropped or coalesced a frame still catches it. Four
// covers a single resolution that reveals more than once (each
// opponent reveals their top card) with room to spare.
//
// The window is ALSO scoped to the current turn, which is the tighter
// of the two bounds in practice and the one that gives "for how long"
// an answer a player can hold in their head: a reveal is on the wire
// for the rest of the turn it happened in. After that the card's
// KnownBy set is the only thing left, which is correct — the table is
// entitled to remember what it saw, and remembering is not the same
// as being shown it again.
const PublicRevealMax = 4

// RevealCardsMax bounds how many cards ONE reveal puts on the wire.
//
// Reveals are usually one card (Dark Confidant) or five (Fact or
// Fiction), and a cap would be pointless for either. It is here for
// Hermit Druid, whose "reveal cards from the top of your library
// until you reveal a basic land card" reveals the WHOLE library when
// there is no basic land in it — the combo the card is played for.
// Without a cap that single activation would put ninety-nine cards on
// a frame already carrying a four-player board.
//
// The cap truncates what is DRAWN, never what is known: RevealView
// .Count reports the true total, and every card's identity is public
// through the ordinary KnownBy path regardless of whether this frame
// found room to name it. Cards are kept from the front, which is the
// order the table saw them in.
//
// Eight is where the budget test put it. A reveal entry is almost all
// strings, and TestRevealFrameWireCost measures a deliberately
// pessimistic one at ~210 B; eight cards times PublicRevealMax is the
// most that fits under revealFrameBudget. It clears every card in the
// catalog that reveals a fixed number — Fact or Fiction's five is the
// largest — so the cap only ever bites on the open-ended ones, which
// are exactly the ones nobody reads card-by-card anyway.
const RevealCardsMax = 8

// RevealView is one reveal: the cards a player showed the whole table
// at one moment, and why.
//
// Every field is either public by construction or a seat index. There
// is no instance ID on this type or on RevealedCardView, by design —
// see the file comment.
type RevealView struct {
	// Seq is the engine sequence number the reveal's first event was
	// stamped with. Monotonic and stable across frames, so a client
	// dedupes and orders on it. It names an event, not a card.
	Seq uint64 `json:"seq"`
	// Turn is the turn sequence the reveal happened on. Present so a
	// reconnecting client can tell a live announcement from one it is
	// seeing for the first time only because it just joined.
	Turn int `json:"turn,omitempty"`
	// Seat is the index of the seat that revealed, or NoSeat.
	Seat int `json:"seat"`
	// Source is the NAME of the card whose effect revealed — never
	// its instance ID. Empty when that card is not somewhere the
	// table can see it.
	Source string `json:"source,omitempty"`
	// Reason is the one-line label the effect wrote, for the banner.
	Reason string `json:"reason,omitempty"`
	// From is the zone the cards were revealed out of ("library",
	// "hand"). The cards did not move: a reveal is not a zone change.
	From string `json:"from,omitempty"`
	// Cards are the revealed cards in the order the table saw them,
	// truncated to RevealCardsMax. Compare against Count.
	Cards []RevealedCardView `json:"cards"`
	// Count is how many cards the reveal actually showed, which is
	// larger than len(Cards) when the cap bit. A client renders the
	// difference as "+N more" rather than pretending it has all of
	// them.
	Count int `json:"count,omitempty"`

	// --- unexported projection inputs, dropped by encoding/json ---

	// cardIDs are the instance IDs the projection resolves printed
	// identity from. They exist for the length of one ViewOfGame call
	// and are never marshalled; the exported half deliberately has no
	// field to put them in.
	cardIDs  []string
	sourceID string
}

// RevealedCardView is one card a reveal showed the table.
//
// Deliberately NOT a CardView. A CardView's first field is its
// instance ID and its redaction story is "hide the printed
// characteristics, keep the handle" — the exact opposite of what a
// reveal needs, which is "publish the printed characteristics, keep
// no handle". Reusing it would have meant remembering to blank the ID
// at every projection site; a type with nowhere to put an ID cannot
// forget.
type RevealedCardView struct {
	Name string `json:"name"`
	// ManaCost is the printed cost string ("{1}{B}"), for the banner
	// and for a client that wants to show what a Dark Confidant flip
	// just cost its controller.
	ManaCost string `json:"mana_cost,omitempty"`
	TypeLine string `json:"type_line,omitempty"`
	// ScryfallID is the printing id the client's image URL builder
	// needs. Public: it identifies a PRINTING, the same way the card's
	// name does, and not this instance of it.
	ScryfallID string `json:"scryfall_id,omitempty"`
}

// publicRevealsOf projects g.Events into the bounded reveal window,
// using v (the already-assembled, unfiltered view) to resolve printed
// identity.
//
// Resolution is NOT gated on the knower set, and that is the
// difference from redactLogForViewer. The log asks "may this viewer
// know what this card is", because a log entry can name a card whose
// identity is still private. A reveal entry can only exist because
// the card's identity was made public, so gating it would be asking a
// question that has already been answered — and answering it wrong:
// the knower set is cleared by a shuffle, so a tutor's reveal would
// silently blank itself the moment the library was shuffled, which is
// the one thing the frame exists to prevent.
//
// Caller must hold the read lock ViewOfGame already holds.
func publicRevealsOf(g *game.Game, v *GameView) []RevealView {
	if g == nil || v == nil {
		return nil
	}
	seatOf := seatIndexer(v)
	turn := 0
	var groups []RevealView
	index := make(map[uint64]int)

	for _, ev := range g.Events {
		if ev.Kind == game.EventStepBegan {
			turn = ev.Amount
			continue
		}
		if ev.Kind != game.EventRevealCards || ev.RevealSeq == 0 {
			continue
		}
		at, ok := index[ev.RevealSeq]
		if !ok {
			groups = append(groups, RevealView{
				Seq:      ev.RevealSeq,
				Turn:     turn,
				Seat:     seatOf(ev.Actor),
				Reason:   ev.Label,
				From:     string(ev.OldZone),
				sourceID: uuidStringOrEmpty(ev.Source),
			})
			at = len(groups) - 1
			index[ev.RevealSeq] = at
		}
		groups[at].cardIDs = append(groups[at].cardIDs, uuidStringOrEmpty(ev.CardID))
	}

	// This turn only, most recent PublicRevealMax. Both bounds are
	// applied from the end, so a turn with a dozen reveals keeps the
	// four the table is most likely still looking at.
	out := make([]RevealView, 0, PublicRevealMax)
	for i := len(groups) - 1; i >= 0 && len(out) < PublicRevealMax; i-- {
		if groups[i].Turn != g.Turn.Seq {
			break
		}
		out = append(out, groups[i])
	}
	if len(out) == 0 {
		return nil
	}
	// Collected newest-first above; the wire wants oldest-first, the
	// same order the log ships in.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	resolveRevealCards(out, v)
	return out
}

// resolveRevealCards fills each entry's printed identity out of the
// assembled view and drops the instance IDs it used to get there.
func resolveRevealCards(entries []RevealView, v *GameView) {
	wanted := make(map[string]bool)
	for _, e := range entries {
		for _, id := range e.cardIDs {
			if id != "" {
				wanted[id] = true
			}
		}
		if e.sourceID != "" {
			wanted[e.sourceID] = true
		}
	}
	if len(wanted) == 0 {
		return
	}

	// Revealed cards are looked up in EVERY zone, including hands and
	// libraries: that is where a revealed card usually still is, and
	// its identity is public precisely because it was revealed.
	all := make(map[string]CardView, len(wanted))
	indexCardsInto(all, wanted, v.Battlefield, v.Stack, v.Exile)
	for _, s := range v.Seats {
		indexCardsInto(all, wanted, s.Hand, s.Library, s.Graveyard, s.Command)
	}
	// The SOURCE is looked up only in zones the table can see. A
	// reveal is caused by a resolving spell or a permanent, so it is
	// always found there in practice; anything else stays anonymous
	// rather than leaking the name of a card in somebody's hand.
	public := make(map[string]CardView, len(entries))
	indexCardsInto(public, wanted, v.Battlefield, v.Stack, v.Exile)
	for _, s := range v.Seats {
		indexCardsInto(public, wanted, s.Graveyard, s.Command)
	}

	for i := range entries {
		e := &entries[i]
		if c, ok := public[e.sourceID]; ok {
			e.Source = c.Name
		}
		e.Count = len(e.cardIDs)
		e.Cards = make([]RevealedCardView, 0, min(len(e.cardIDs), RevealCardsMax))
		for _, id := range e.cardIDs {
			if len(e.Cards) >= RevealCardsMax {
				break
			}
			c, ok := all[id]
			if !ok {
				// The card has left every tracked zone since the
				// reveal. Nothing honest to say about it, and a
				// placeholder would be worse than one fewer card.
				continue
			}
			e.Cards = append(e.Cards, RevealedCardView{
				Name:       c.Name,
				ManaCost:   c.ManaCost,
				TypeLine:   c.TypeLine,
				ScryfallID: c.ScryfallID,
			})
		}
		e.cardIDs = nil
		e.sourceID = ""
	}
}
