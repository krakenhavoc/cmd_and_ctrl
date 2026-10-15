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
// with this text goes through QueueAnyPlayerPaysForEffect.

// EventAnyPlayerPaid records that Actor paid an any-player offer
// (ADR 0150 §5): Source is the card whose effect it was, Label the
// payment's words ("{2}"). Public: the table reads back who paid.
const EventAnyPlayerPaid EventKind = "any_player_paid"

// EventAnyPlayerDeclined records that Actor declined an any-player
// offer, or said "pay" and could not (CR 118.3). One per decline, in
// the order asked. Source and Label as for EventAnyPlayerPaid.
const EventAnyPlayerDeclined EventKind = "any_player_declined"

// AnyPlayerPays is the whole of an any-player offer.
type AnyPlayerPays struct {
	// Source is the card whose effect asks; Controller is that
	// effect's controller ("you" on the card). The view names the
	// offer as Controller's Source, and a bot never pays its own.
	Source     uuid.UUID
	Controller uuid.UUID
	// Cost is the mana payment, as ParseCost reads it, with any X
	// already worked out by the card (Soul Strings' {X}).
	Cost string
	// Action is a payment that is not mana (ADR 0108 §5, ADR 0150 §3):
	// life, damage, a sacrifice, a discard, or nothing at all. Nil for
	// a mana payment; when set, Cost is the payment's printed words.
	Action *PayAction
	// PerPayer builds the payment for the player being asked, when it
	// differs from player to player: Temporal Extortion's "half their
	// life, rounded up" (ADR 0150 §3). It runs as each link is queued,
	// so it reads that player's life then, and returns the payment and
	// its words, which replace Action and Cost for that link. Optional.
	PerPayer func(g *Game, payer uuid.UUID) (PayAction, string)
	// AcceptLabel and DeclineLabel name the two answers when "Pay" and
	// "Don't pay" read wrong: "Take 5 damage", "Let them have it".
	// Optional; empty keeps the client's words.
	AcceptLabel, DeclineLabel string
	// Question is the dialog header each offeree sees.
	Question string
	// Opponents narrows the offer to the controller's opponents ("any
	// opponent may", Innocent Traveler). False is "any player".
	Opponents bool
	// HelpsController marks an offer whose payment is what the
	// controller wants: Icy Prison's "sacrifice this unless any player
	// pays {3}", where paying keeps their permanent. False is the rest
	// of the family, where paying stops the controller's effect. Plain
	// data for the view's words and the bot's answer.
	HelpsController bool
	// OnPaid is "If a player does, …", handed the payer. Optional.
	OnPaid func(g *Game, payer uuid.UUID) error
	// OnNobody is "unless …" / "If no one does, …". It runs exactly
	// once, after the last player has declined. Optional.
	OnNobody func(g *Game) error
	// Guards is the stack object the outcome decides (Dash Hopes'
	// "counter it"), held on the stack while the offer is open, or
	// uuid.Nil.
	Guards uuid.UUID
}

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
	// HelpsController is AnyPlayerPays.HelpsController.
	HelpsController bool `json:"helpsController,omitempty"`
	// EachLife is the life one object costs on a per-object offer
	// (ADR 0187 §2, AnyPlayerPaysForEach): the leg is a choose_cards
	// whose answer is the objects the chooser will pay that much for.
	// Zero on every other offer.
	EachLife int `json:"eachLife,omitempty"`
}

func cloneAnyPlayerOffer(o *AnyPlayerOffer) *AnyPlayerOffer {
	if o == nil {
		return nil
	}
	out := *o
	out.Order = copyUUIDs(o.Order)
	out.Declined = copyUUIDs(o.Declined)
	return &out
}

// QueueAnyPlayerPaysForEffect offers p's payment to every player still
// in the game in APNAP order from the active player, one pay-unless at
// a time, and runs OnPaid for the first who pays or OnNobody when all
// decline. With nobody to ask (every player gone, or no opponent left
// for an Opponents offer) OnNobody runs at once.
//
// Caller must hold g.mu.
func (g *Game) QueueAnyPlayerPaysForEffect(p AnyPlayerPays) error {
	if p.Action != nil {
		a := *p.Action
		p.Action = &a
	} else if _, err := ParseCost(p.Cost); err != nil && p.PerPayer == nil {
		// (A PerPayer offer builds each link's payment as it is
		// queued, so its Cost is not mana.)
		// A programming error, surfaced once rather than once per
		// player; the card still does something.
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   p.Source,
			ErrorMsg: "any player pays: unparseable cost " + p.Cost + ": " + err.Error(),
		})
		if p.OnNobody != nil {
			return p.OnNobody(g)
		}
		return nil
	}
	order := g.votersFromLocked(g.activePlayerIDLocked())
	if p.Opponents {
		kept := order[:0]
		for _, id := range order {
			if id != p.Controller {
				kept = append(kept, id)
			}
		}
		order = kept
	}
	offer := &AnyPlayerOffer{
		Controller: p.Controller,
		Order:      order,
		Opponents:  p.Opponents,

		HelpsController: p.HelpsController,
	}
	return g.offerAnyPlayerFromLocked(p, offer, 0, g.currentTurnStepLocked())
}

// offerAnyPlayerFromLocked asks Order[next], or the first player after
// it still in the game, and runs OnNobody when there is none. `owed` is
// the step the offer was made in: every link is owed in it. Caller must
// hold g.mu.
func (g *Game) offerAnyPlayerFromLocked(p AnyPlayerPays, offer *AnyPlayerOffer, next int, owed TurnStep) error {
	for ; next < len(offer.Order); next++ {
		who := offer.Order[next]
		if !g.playerInGameLocked(who) {
			// CR 800.4f: a player who has left is not asked.
			continue
		}
		link := cloneAnyPlayerOffer(offer)
		after := next + 1
		cost, action := p.Cost, p.Action
		if p.PerPayer != nil {
			// ADR 0150 §3: this player's own price, worked out now.
			a, words := p.PerPayer(g, who)
			action, cost = &a, words
		}
		decline := func(g *Game) error {
			g.EmitEvent(Event{Kind: EventAnyPlayerDeclined, Actor: who, Source: p.Source, Label: cost})
			moved := cloneAnyPlayerOffer(link)
			moved.Declined = append(moved.Declined, who)
			return g.offerAnyPlayerFromLocked(p, moved, after, owed)
		}
		pay := func(g *Game) error {
			g.EmitEvent(Event{Kind: EventAnyPlayerPaid, Actor: who, Source: p.Source, Label: cost})
			if p.OnPaid == nil {
				return nil
			}
			return p.OnPaid(g, who)
		}
		return g.queuePayUnlessLocked(who, p.Source, cost, p.Question, decline,
			owed, p.Guards, nil,
			payUnlessExtras{
				onPay: pay, action: action, offer: link,
				acceptLabel: p.AcceptLabel, declineLabel: p.DeclineLabel,
			})
	}
	if p.OnNobody == nil {
		return nil
	}
	return p.OnNobody(g)
}
