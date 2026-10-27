package game

import "github.com/google/uuid"

// any_player_pays.go — "unless any player pays {N}" and "any player may
// pay {N}. If no one does, …" (ADR 0150 decisions 1 and 2, #1903):
// one payment offered to the whole table.
//
// THE SHAPE. The offer is a chain of ordinary pay-unless prompts, one
// at a time. Every player still in the game is listed in APNAP order
// from the ACTIVE player (CR 101.4, owner decision 1), the controller
// included ("any player" includes them; "any opponent" does not). The
// first is asked; a decline asks the next; the first payer ends the
// offer and nobody after them is asked; when everyone has declined the
// effect happens. Each link is a PendingChoicePayUnless built by
// queuePayUnlessLocked, so its validator, its payment, the enumerator's
// moves, the dock's prompt and the departure table's decline all answer
// it unchanged.
//
// WHERE IT DIFFERS FROM A LITERAL CR 101.4. The rule has every player
// choose in turn and then "the actions happen simultaneously", so a
// later player could also pay after an earlier one already had. On
// every card of the family a second payment buys nothing, so the engine
// pays each answer as it is given and stops at the first payer (ADR
// 0150 decision 1's recorded simplification).
//
// IT BLOCKS ITS STEP (decision 2, owner decision 2). The effect is
// paused partway through resolving — Rhystic Shield's +0/+2 must land
// before damage — so every link is owed in the step the offer was made
// in (PendingChoice.OwedInStep) and, when the outcome decides a spell's
// fate, guards that spell (GuardsStackItem). A step-owed prompt takes
// no standing answer (payUnlessCovered), so a remembered "always pay
// Rhystic Study" never pays an opponent's Rhystic Tutor.
//
// ONE DOOR. A card that queued a bare pay-unless to each player would
// lose the halt and the standing-answer exclusion together; every card
// with this text goes through QueueAnyPlayerPaysKeyedForEffect
// (any_player_keyed.go), whose consequences are a registered key. (The
// closure form, QueueAnyPlayerPaysForEffect with its AnyPlayerPays
// struct, was deleted in #4253's last part once only tests called it.)

// EventAnyPlayerPaid records that Actor paid an any-player offer
// (ADR 0150 §5): Source is the card whose effect it was, Label the
// payment's words ("{2}"). Public: the table reads back who paid.
const EventAnyPlayerPaid EventKind = "any_player_paid"

// EventAnyPlayerDeclined records that Actor declined an any-player
// offer, or said "pay" and could not (CR 118.3). One per decline, in
// the order asked. Source and Label as for EventAnyPlayerPaid.
const EventAnyPlayerDeclined EventKind = "any_player_declined"

// AnyPlayerOffer is an any-player offer in progress, as plain data on
// the open link (PendingChoice.AnyPlayerOffer): whose effect it is,
// everyone it may be offered to in the order they are asked, and who
// has declined so far. Read by the view (every seat sees who is being
// asked) and by the bot. Each link carries its own copy.
type AnyPlayerOffer struct {
	Controller uuid.UUID   `json:"controller"`
	Order      []uuid.UUID `json:"order"`
	Declined   []uuid.UUID `json:"declined,omitempty"`
	Opponents  bool        `json:"opponents,omitempty"`
	// HelpsController is AnyPlayerPaysKeyed.HelpsController.
	HelpsController bool `json:"helpsController,omitempty"`
	// EachLife is the life one object costs on a per-object offer
	// (ADR 0187 §2, AnyPlayerPaysForEach): the leg is a choose_cards
	// whose answer is the objects the chooser will pay that much for.
	// Zero on every other offer.
	EachLife int `json:"eachLife,omitempty"`
	// Objects, Picks and Noun are a KEYED per-object offer's state
	// (#4264, any_player_each_keyed.go): the objects it is about, each
	// with the epoch it had when the offer was made; the picks of the
	// seats already asked, in order; and the object's noun ("land"). They
	// are what the closure form's ordered collector held in its frames.
	// Additive within the version; empty on every other offer.
	Objects []ObjectRef `json:"objects,omitempty"`
	Picks   []OfferPick `json:"picks,omitempty"`
	Noun    string      `json:"noun,omitempty"`
}

// OfferPick is one seat's answer to a keyed per-object offer's leg: the
// objects it will pay for. A seat that chose nothing has an empty list.
type OfferPick struct {
	Seat  uuid.UUID   `json:"seat"`
	Cards []uuid.UUID `json:"cards,omitempty"`
}

func cloneAnyPlayerOffer(o *AnyPlayerOffer) *AnyPlayerOffer {
	if o == nil {
		return nil
	}
	out := *o
	out.Order = copyUUIDs(o.Order)
	out.Declined = copyUUIDs(o.Declined)
	out.Objects = cloneObjectRefs(o.Objects)
	if len(o.Picks) > 0 {
		out.Picks = make([]OfferPick, len(o.Picks))
		for i, p := range o.Picks {
			out.Picks[i] = OfferPick{Seat: p.Seat, Cards: copyUUIDs(p.Cards)}
		}
	}
	return &out
}
