package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// any_player_keyed.go — an any-player offer ("any player may pay {N}. If
// a player does, ...; if no one does, ...") whose consequences are a KEY
// rather than closures (#4264), so a table waiting on any link of the
// offer is a restore point.
//
// The offer is a chain of ordinary pay-unless prompts, one per player
// (any_player_pays.go), and every link's decline is a closure over the
// whole offer: who is next, the cost, the labels, the two consequences.
// None of that is state a restore point can carry as a func, but all of
// it is data a link already has or that rides on it: the payer (Chooser),
// the card (Source), the offer (AnyPlayerOffer: the controller, the order
// everyone is asked in, who has declined, the two flags), the cost as
// printed (PayCost) or a non-mana payment (PayActionClause), the labels,
// the step the offer is owed in (OwedInStep), the spell it guards
// (GuardsStackItem), and the key and the IDs the card put there
// (PickThen, OptionCarry). The next link is the one after the payer in
// the offer's order, so a link rebuilds its own decline from its data
// (anyPlayerFrameOf, read by payUnlessFrameOf), and every path that
// answers a pay-unless answers a keyed leg the same way.
//
// The key names the two consequences and the optional per-payer payment
// (RegisterAnyPlayerPays). It shares the effect-key namespace and the
// ledger ("anyplayer <key>"). A binary from before this change refuses a
// restore point naming a key it does not have (checkEffectKeys).
//
// The per-object offer (QueueAnyPlayerPaysForEachForEffect, whose legs
// are choose_cards prompts) is not keyed here.

// AnyPlayerAnswer is what a keyed any-player consequence is handed.
type AnyPlayerAnswer struct {
	// Source is the card whose effect it was and Controller that
	// effect's controller ("you" on the card).
	Source, Controller uuid.UUID
	// Payer is the player who paid, for the "if a player does" branch;
	// uuid.Nil for the "if no one does" branch.
	Payer uuid.UUID
	// Carry is the plain data the queuing card put on the offer.
	Carry []uuid.UUID
	// Guards is the stack object the offer guarded, or uuid.Nil.
	Guards uuid.UUID
	// Item is the stack item that asked, rebuilt from the data the offer
	// carried (AskedItem), or nil when the card put none there. See
	// PayUnlessAnswer.Item.
	Item *StackItem
}

// AnyPlayerFunc is a registered consequence. It runs with g.mu held, may
// queue prompts of its own, and must capture nothing.
type AnyPlayerFunc func(g *Game, r AnyPlayerAnswer) error

// AnyPlayerPerPayerFunc is a registered per-payer payment: the payment
// and its words for the player being asked, worked out as their link is
// queued (Temporal Extortion's half their life, rounded up). It reads
// the game and the carried IDs and nothing else.
type AnyPlayerPerPayerFunc func(g *Game, payer uuid.UUID, carry []uuid.UUID) (PayAction, string)

// AnyPlayerThen names a card's registered any-player consequences. The
// key is unexported, so the only way to hold one is RegisterAnyPlayerPays.
type AnyPlayerThen struct{ key string }

// Key is the consequences' on-disk key.
func (t AnyPlayerThen) Key() string { return t.key }

// anyPlayerEntry is one registered key's consequences; any may be nil.
type anyPlayerEntry struct {
	onPaid, onNobody AnyPlayerFunc
	perPayer         AnyPlayerPerPayerFunc
}

// anyPlayerThens is guarded by effectRegistryMu, with the other effect
// keys, because the keys share one namespace and one ledger.
var anyPlayerThens = map[string]anyPlayerEntry{}

// RegisterAnyPlayerPays registers a keyed any-player offer's
// consequences: `onPaid` is "if a player does, ..." (handed the payer),
// `onNobody` "if no one does, ..." and `perPayer` the payment worked out
// for each player. Any may be nil, but not all three. Call it once, from
// a package-level var. Panics on a bad or duplicate key. The key is an
// on-disk identity: never renamed, never reused.
func RegisterAnyPlayerPays(key string, onPaid, onNobody AnyPlayerFunc, perPayer AnyPlayerPerPayerFunc) AnyPlayerThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("any-player consequences", key)
	if onPaid == nil && onNobody == nil && perPayer == nil {
		panic(fmt.Sprintf("game: any-player consequences %q have no function", key))
	}
	anyPlayerThens[key] = anyPlayerEntry{onPaid: onPaid, onNobody: onNobody, perPayer: perPayer}
	return AnyPlayerThen{key: key}
}

func lookupAnyPlayerPays(key string) (anyPlayerEntry, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	e, ok := anyPlayerThens[key]
	return e, ok
}

// KnownAnyPlayerPays reports whether this binary has consequences
// registered under key.
func KnownAnyPlayerPays(key string) bool { _, ok := lookupAnyPlayerPays(key); return ok }

// registeredAnyPlayerKeysLocked lists the ledger lines, "anyplayer
// <key>". Caller holds effectRegistryMu.
func registeredAnyPlayerKeysLocked() []string {
	out := make([]string, 0, len(anyPlayerThens))
	for k := range anyPlayerThens {
		out = append(out, "anyplayer "+k)
	}
	sort.Strings(out)
	return out
}

// AnyPlayerPaysKeyed is the whole of a keyed any-player offer:
// AnyPlayerPays with the closures replaced by Then and Carry.
type AnyPlayerPaysKeyed struct {
	// Source is the card whose effect asks; Controller is that effect's
	// controller.
	Source, Controller uuid.UUID
	// Cost is the mana payment, as ParseCost reads it, or, with an
	// Action, the payment's printed words.
	Cost string
	// Action is a payment that is not mana (ADR 0108 §5, ADR 0150 §3).
	Action *PayAction
	// AcceptLabel and DeclineLabel name the two answers; Question is the
	// dialog header each offeree sees.
	AcceptLabel, DeclineLabel, Question string
	// Opponents narrows the offer to the controller's opponents.
	Opponents bool
	// HelpsController is AnyPlayerPays.HelpsController.
	HelpsController bool
	// Guards is the stack object the outcome decides, held on the stack
	// while the offer is open, or uuid.Nil.
	Guards uuid.UUID
	// Then names the consequences (RegisterAnyPlayerPays); the
	// consequences are handed Carry.
	Then  AnyPlayerThen
	Carry []uuid.UUID
	// Asked is the asking stack item as data (NewAskedItem), handed to the
	// consequences as AnyPlayerAnswer.Item (#4264). Nil when they read
	// none.
	Asked *AskedItem
}

// QueueAnyPlayerPaysKeyedForEffect is QueueAnyPlayerPaysForEffect with
// the consequences as a registered key. Every link is a prompt that
// carries no closure, so a table waiting on any of them is a restore
// point. With nobody to ask, the "if no one does" branch runs at once.
//
// Caller must hold g.mu.
func (g *Game) QueueAnyPlayerPaysKeyedForEffect(p AnyPlayerPaysKeyed) error {
	if p.Then.key == "" {
		return fmt.Errorf("%w: keyed any-player offer with no consequences", ErrInvalidParam)
	}
	entry, known := lookupAnyPlayerPays(p.Then.key)
	if !known {
		return fmt.Errorf("%w: any-player consequences %q", ErrUnknownEffectKey, p.Then.key)
	}
	p.Action = copyPayAction(p.Action)
	p.Carry = copyUUIDs(p.Carry)
	if p.Action == nil && entry.perPayer == nil {
		if _, err := ParseCost(p.Cost); err != nil {
			// A programming error, surfaced once rather than once per
			// player; the card still does something.
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Source:   p.Source,
				ErrorMsg: "any player pays: unparseable cost " + p.Cost + ": " + err.Error(),
			})
			return g.runAnyPlayerNobodyLocked(p)
		}
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
		Controller:      p.Controller,
		Order:           order,
		Opponents:       p.Opponents,
		HelpsController: p.HelpsController,
	}
	return g.offerAnyPlayerKeyedFromLocked(p, offer, 0, g.currentTurnStepLocked())
}

// runAnyPlayerNobodyLocked is the "if no one does" branch of a keyed
// offer. Caller must hold g.mu.
func (g *Game) runAnyPlayerNobodyLocked(p AnyPlayerPaysKeyed) error {
	entry, _ := lookupAnyPlayerPays(p.Then.key)
	if entry.onNobody == nil {
		return nil
	}
	return entry.onNobody(g, AnyPlayerAnswer{
		Source: p.Source, Controller: p.Controller, Carry: copyUUIDs(p.Carry), Guards: p.Guards,
		Item: p.Asked.StackItem(),
	})
}

// offerAnyPlayerKeyedFromLocked asks Order[next], or the first player
// after it still in the game, and runs the nobody branch when there is
// none. `owed` is the step the offer was made in. Caller must hold g.mu.
func (g *Game) offerAnyPlayerKeyedFromLocked(p AnyPlayerPaysKeyed, offer *AnyPlayerOffer, next int, owed TurnStep) error {
	entry, _ := lookupAnyPlayerPays(p.Then.key)
	for ; next < len(offer.Order); next++ {
		who := offer.Order[next]
		if !g.playerInGameLocked(who) {
			// CR 800.4f: a player who has left is not asked.
			continue
		}
		cost, action := p.Cost, p.Action
		if entry.perPayer != nil {
			// ADR 0150 §3: this player's own price, worked out now.
			a, words := entry.perPayer(g, who, p.Carry)
			action, cost = &a, words
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:            PendingChoicePayUnless,
			Chooser:         who,
			Count:           1,
			Source:          p.Source,
			Reason:          p.Question,
			PayCost:         cost,
			OwedInStep:      owed,
			GuardsStackItem: p.Guards,
			AnyPlayerOffer:  cloneAnyPlayerOffer(offer),
			AcceptLabel:     p.AcceptLabel,
			DeclineLabel:    p.DeclineLabel,
			PayActionClause: copyPayAction(action),
			Asked:           copyAskedItem(p.Asked),
			PickThen:        p.Then.key,
			OptionCarry:     copyUUIDs(p.Carry),
		})
		return nil
	}
	return g.runAnyPlayerNobodyLocked(p)
}

// anyPlayerFrameOf is the frame a keyed any-player leg answers through,
// built from the leg's data (see the file comment), or nil when `c` is
// not one. The decline records the refusal and asks the player after
// this one; the payment records it and runs the "if a player does"
// branch.
func anyPlayerFrameOf(c *PendingChoice) *payUnlessFrame {
	if c == nil || c.Kind != PendingChoicePayUnless || c.PickThen == "" || c.AnyPlayerOffer == nil {
		return nil
	}
	entry, known := lookupAnyPlayerPays(c.PickThen)
	if !known {
		return nil
	}
	f := &payUnlessFrame{}
	if c.PayActionClause != nil {
		f.action = copyPayAction(c.PayActionClause)
	} else {
		parsed, err := ParseCost(c.PayCost)
		if err != nil {
			return nil
		}
		f.cost = parsed
	}
	who, source, words := c.Chooser, c.Source, c.PayCost
	offer := cloneAnyPlayerOffer(c.AnyPlayerOffer)
	p := AnyPlayerPaysKeyed{
		Source:          c.Source,
		Controller:      offer.Controller,
		Cost:            c.PayCost,
		Action:          copyPayAction(c.PayActionClause),
		AcceptLabel:     c.AcceptLabel,
		DeclineLabel:    c.DeclineLabel,
		Question:        c.Reason,
		Opponents:       offer.Opponents,
		HelpsController: offer.HelpsController,
		Guards:          c.GuardsStackItem,
		Then:            AnyPlayerThen{key: c.PickThen},
		Carry:           copyUUIDs(c.OptionCarry),
		Asked:           copyAskedItem(c.Asked),
	}
	owed := c.OwedInStep
	after := len(offer.Order)
	for i, id := range offer.Order {
		if id == who {
			after = i + 1
			break
		}
	}
	f.onDecline = func(g *Game) error {
		g.EmitEvent(Event{Kind: EventAnyPlayerDeclined, Actor: who, Source: source, Label: words})
		moved := cloneAnyPlayerOffer(offer)
		moved.Declined = append(moved.Declined, who)
		return g.offerAnyPlayerKeyedFromLocked(p, moved, after, owed)
	}
	f.onPay = func(g *Game) error {
		g.EmitEvent(Event{Kind: EventAnyPlayerPaid, Actor: who, Source: source, Label: words})
		if entry.onPaid == nil {
			return nil
		}
		return entry.onPaid(g, AnyPlayerAnswer{
			Source: p.Source, Controller: p.Controller, Payer: who, Carry: copyUUIDs(p.Carry), Guards: p.Guards,
			Item: p.Asked.StackItem(),
		})
	}
	return f
}
