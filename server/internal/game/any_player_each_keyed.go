package game

import (
	"fmt"

	"github.com/google/uuid"
)

// any_player_each_keyed.go — the per-object any-player offer ("for each
// <object>, <effect> unless any player pays <N life>", ADR 0187 §2) whose
// legs and whose consequence are plain data and a KEY rather than closures
// (#4264), so a table waiting on any seat's leg is a restore point.
//
// The closure form (any_player_pays_each.go) collects the seats' picks
// with ADR 0175's ordered collector, whose every leg's continuation is a
// closure over the legs still to come, the picks so far, the objects'
// epochs and the card's OnUnsaved. All of that is data a leg can carry:
//
//   - the objects, each with the epoch it had when the offer was made, and
//     the picks of the seats already asked, ride the offer
//     (AnyPlayerOffer.Objects, .Picks), which every leg already carries;
//   - the seat after this one is read off the offer's order, with the same
//     test the closure form used to make the legs (a seat in the game that
//     can pay for something), because no life moves before the payment
//     phase and so the figure still holds when that seat is reached;
//   - the question, the card, the carried IDs and the asking item ride the
//     leg itself (Reason, Source, OptionCarry, Asked);
//   - the key is the leg's continuation (PickThen). RegisterAnyPlayerPaysEach
//     registers the card's "unless" consequence as a keyed choose_cards
//     continuation under it, and the engine's leg step wraps it, so the key
//     shares the effect-key namespace and the ledger ("choose <key>"), and
//     a binary that lacks it refuses a restore point naming it
//     (checkEffectKeys).
//
// Everything else is the closure form's: the seats are asked in APNAP
// order, each seeing the earlier seats' picks (CR 101.4b); a seat that has
// left pays for nothing (CR 800.4f); then every payment is made, one life
// payment per object, and the consequence gets the objects nobody paid for
// that are still the same objects (CR 400.7).

// AnyPlayerEachAnswer is what a keyed per-object consequence is handed.
type AnyPlayerEachAnswer struct {
	// Source is the card whose effect it was and Controller that
	// effect's controller ("you" on the card).
	Source, Controller uuid.UUID
	// Unsaved are the objects nobody paid for that are still the same
	// objects.
	Unsaved []uuid.UUID
	// Carry is the plain data the queuing card put on the offer.
	Carry []uuid.UUID
	// Item is the stack item that asked, rebuilt from the data the offer
	// carried, or nil when the card put none there.
	Item *StackItem
}

// AnyPlayerEachFunc is a registered consequence. It runs with g.mu held,
// may queue prompts of its own, and must capture nothing.
type AnyPlayerEachFunc func(g *Game, r AnyPlayerEachAnswer) error

// AnyPlayerEachThen names a card's registered per-object consequence. The
// key is unexported, so the only way to hold one is
// RegisterAnyPlayerPaysEach.
type AnyPlayerEachThen struct{ key string }

// Key is the consequence's on-disk key.
func (t AnyPlayerEachThen) Key() string { return t.key }

// RegisterAnyPlayerPaysEach registers the consequence of a keyed
// per-object offer: `fn` is "the objects nobody paid for ...". Call it
// once, from a package-level var. Panics on a bad or duplicate key. The
// key is an on-disk identity: never renamed, never reused. It is ledgered
// as the offer's choose_cards continuation ("choose <key>").
func RegisterAnyPlayerPaysEach(key string, fn AnyPlayerEachFunc) AnyPlayerEachThen {
	if fn == nil {
		panic(fmt.Sprintf("game: per-object any-player consequence %q has no function", key))
	}
	RegisterChooseCardsThen(key, func(g *Game, r CardsChosen) error {
		return g.anyPlayerEachLegAnsweredLocked(r, fn)
	})
	effectRegistryMu.Lock()
	anyPlayerEachFuncs[key] = fn
	effectRegistryMu.Unlock()
	return AnyPlayerEachThen{key: key}
}

// anyPlayerEachFuncs are the consequences, by key, for the offer that
// needs one before any leg has been asked. Guarded by effectRegistryMu.
var anyPlayerEachFuncs = map[string]AnyPlayerEachFunc{}

func lookupAnyPlayerEachFunc(key string) (AnyPlayerEachFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := anyPlayerEachFuncs[key]
	return fn, ok
}

// eachOfferState is what a leg's answer needs beside the offer: the
// consequence, and the card's and the prompt's own data.
type eachOfferState struct {
	key      string
	fn       AnyPlayerEachFunc
	source   uuid.UUID
	question string
	carry    []uuid.UUID
	asked    *AskedItem
}

// anyPlayerEachAffordableLocked is the most payments `who` can make one
// after another, 0 when they can pay for nothing. CR 119.4 applies to each
// payment and CR 119.8 to all of them.
func (g *Game) anyPlayerEachAffordableLocked(who uuid.UUID, eachLife, objects int) int {
	pl := g.playerByIDLocked(who)
	if pl == nil || pl.Eliminated || eachLife <= 0 || !g.CanPayLifeLocked(pl, eachLife) {
		return 0
	}
	return min(pl.Life/eachLife, objects)
}

// queueAnyPlayerPaysForEachKeyedLocked is the door's keyed branch:
// QueueAnyPlayerPaysForEachForEffect has frozen the objects with the epoch
// each has now, and the closure form's two early outs apply here too (a
// payment that is not a payment, and nothing to pay for): the consequence
// runs at once on every object. Caller must hold g.mu.
func (g *Game) queueAnyPlayerPaysForEachKeyedLocked(p AnyPlayerPaysForEach, objects []uuid.UUID, epochs map[uuid.UUID]int) error {
	refs := make([]ObjectRef, 0, len(objects))
	for _, id := range objects {
		refs = append(refs, ObjectRef{ID: id, Epoch: epochs[id]})
	}
	if p.EachLife <= 0 {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   p.Source,
			ErrorMsg: fmt.Sprintf("any player pays for each: non-positive life payment %d", p.EachLife),
		})
		return g.settleAnyPlayerEachWithoutAsking(p, refs)
	}
	if len(refs) == 0 {
		return g.settleAnyPlayerEachWithoutAsking(p, refs)
	}
	question := p.Question
	if question == "" {
		question = fmt.Sprintf("Choose which %ss to pay %d life for", p.Noun, p.EachLife)
	}
	return g.queueAnyPlayerEachKeyedLocked(p, refs, question)
}

// queueAnyPlayerEachKeyedLocked is the keyed form of the offer, for the
// objects frozen in `refs`, each with the epoch it had. Caller must hold
// g.mu.
func (g *Game) queueAnyPlayerEachKeyedLocked(p AnyPlayerPaysForEach, refs []ObjectRef, question string) error {
	fn, known := lookupAnyPlayerEachFunc(p.Then.key)
	if !known {
		return fmt.Errorf("%w: per-object any-player consequence %q", ErrUnknownEffectKey, p.Then.key)
	}
	words := fmt.Sprintf("%d life per %s", p.EachLife, p.Noun)
	order := g.votersFromLocked(g.activePlayerIDLocked())
	offer := &AnyPlayerOffer{
		Controller: p.Controller,
		Order:      order,
		EachLife:   p.EachLife,
		Objects:    cloneObjectRefs(refs),
		Noun:       p.Noun,
	}
	for _, who := range order {
		if g.anyPlayerEachAffordableLocked(who, p.EachLife, len(refs)) == 0 {
			// A prompt whose only answer is "none" is a wasted click.
			g.EmitEvent(Event{Kind: EventAnyPlayerDeclined, Actor: who, Source: p.Source, Label: words})
		}
	}
	st := eachOfferState{
		key: p.Then.key, fn: fn, source: p.Source, question: question,
		carry: copyUUIDs(p.Carry), asked: copyAskedItem(p.Asked),
	}
	return g.queueNextAnyPlayerEachLegLocked(st, offer, 0)
}

// anyPlayerEachLegAnsweredLocked is the registered continuation of every
// leg: it files the answer on the offer and asks the next seat, or settles
// the offer after the last. Caller must hold g.mu.
func (g *Game) anyPlayerEachLegAnsweredLocked(r CardsChosen, fn AnyPlayerEachFunc) error {
	offer := cloneAnyPlayerOffer(r.Offer)
	if offer == nil {
		return fmt.Errorf("%w: a per-object any-player leg with no offer", ErrInvalidParam)
	}
	var picked []uuid.UUID
	if r.Answered {
		picked = copyUUIDs(r.Picked)
	}
	offer.Picks = append(offer.Picks, OfferPick{Seat: r.Chooser, Cards: picked})
	st := eachOfferState{key: r.Key, fn: fn, source: r.Source, question: r.Question, carry: r.Carry, asked: NewAskedItem(r.Asked)}
	next := len(offer.Order)
	for i, id := range offer.Order {
		if id == r.Chooser {
			next = i + 1
			break
		}
	}
	return g.queueNextAnyPlayerEachLegLocked(st, offer, next)
}

// queueNextAnyPlayerEachLegLocked asks the first seat at or after
// offer.Order[from] that has a leg, and settles the offer when there is
// none. A seat has a leg when it is in the game, can pay for something and
// has an object left to pay for. Caller must hold g.mu.
func (g *Game) queueNextAnyPlayerEachLegLocked(st eachOfferState, offer *AnyPlayerOffer, from int) error {
	for i := from; i < len(offer.Order); i++ {
		who := offer.Order[i]
		affordable := g.anyPlayerEachAffordableLocked(who, offer.EachLife, len(offer.Objects))
		if affordable == 0 {
			continue
		}
		var cards []uuid.UUID
		for _, ref := range offer.Objects {
			if c := findBattlefieldCard(g, ref.ID); c != nil && c.ObjectEpoch == ref.Epoch {
				cards = append(cards, ref.ID)
			}
		}
		if len(cards) == 0 {
			continue
		}
		hi := min(affordable, len(cards))
		earlier := make(PromptedPicks, 0, len(offer.Picks))
		for _, pick := range offer.Picks {
			earlier = append(earlier, SeatCards{Seat: pick.Seat, Cards: copyUUIDs(pick.Cards)})
		}
		leg := cloneAnyPlayerOffer(offer)
		leg.Declined = g.anyPlayerEachDeclinedBeforeLocked(offer, i)
		g.QueueChoiceForEffect(PendingChoice{
			Kind:            PendingChoiceChooseCards,
			Chooser:         who,
			FromPlayer:      who,
			Count:           hi,
			Source:          st.source,
			Reason:          st.question,
			ChooseCards:     cards,
			ChooseMin:       0,
			ChooseMax:       hi,
			ChooseSuggested: g.suggestUnchosenOwnLocked(who, cards, hi, earlier),
			AnyPlayerOffer:  leg,
			PickThen:        st.key,
			OptionCarry:     copyUUIDs(st.carry),
			Asked:           copyAskedItem(st.asked),
		})
		return nil
	}
	return g.settleAnyPlayerEachLocked(st, offer)
}

// anyPlayerEachDeclinedBeforeLocked are the seats before offer.Order[at]
// that chose nothing: those with no leg at all (they could pay for
// nothing) and those whose answer was empty. Caller must hold g.mu.
func (g *Game) anyPlayerEachDeclinedBeforeLocked(offer *AnyPlayerOffer, at int) []uuid.UUID {
	var out []uuid.UUID
	for _, who := range offer.Order[:at] {
		picked := false
		answered := false
		for _, pick := range offer.Picks {
			if pick.Seat == who {
				answered = true
				picked = len(pick.Cards) > 0
			}
		}
		if answered && picked {
			continue
		}
		out = append(out, who)
	}
	return out
}

// settleAnyPlayerEachLocked is the payment phase: every seat, in order,
// pays EachLife once per object it chose that is still the same object, as
// separate payments (owner decision 2), and then the consequence gets the
// objects nobody paid for. Caller must hold g.mu.
func (g *Game) settleAnyPlayerEachLocked(st eachOfferState, offer *AnyPlayerOffer) error {
	words := fmt.Sprintf("%d life per %s", offer.EachLife, offer.Noun)
	epochs := make(map[uuid.UUID]int, len(offer.Objects))
	for _, ref := range offer.Objects {
		epochs[ref.ID] = ref.Epoch
	}
	same := func(id uuid.UUID) bool {
		c := findBattlefieldCard(g, id)
		epoch, ok := epochs[id]
		return c != nil && ok && c.ObjectEpoch == epoch
	}
	saved := make(map[uuid.UUID]bool)
	for _, entry := range offer.Picks {
		paid := 0
		for _, id := range entry.Cards {
			if !same(id) {
				continue
			}
			pl := g.playerByIDLocked(entry.Seat)
			// CR 800.4f: a seat that has left pays nothing. CR 118.3
			// and 119.4: nor one that can't pay now.
			if pl == nil || pl.Eliminated || !g.CanPayLifeLocked(pl, offer.EachLife) {
				continue
			}
			if err := g.payLifeAsCostLocked(st.source, entry.Seat, offer.EachLife); err != nil {
				continue
			}
			paid++
			saved[id] = true
		}
		if paid == 0 {
			g.EmitEvent(Event{Kind: EventAnyPlayerDeclined, Actor: entry.Seat, Source: st.source, Label: words})
			continue
		}
		noun := offer.Noun
		if paid != 1 {
			noun += "s"
		}
		g.EmitEvent(Event{
			Kind: EventAnyPlayerPaid, Actor: entry.Seat, Source: st.source,
			Label: fmt.Sprintf("%d life (%d %s)", paid*offer.EachLife, paid, noun),
		})
	}
	fn := st.fn
	if fn == nil {
		return nil
	}
	var unsaved []uuid.UUID
	for _, ref := range offer.Objects {
		if !saved[ref.ID] && same(ref.ID) {
			unsaved = append(unsaved, ref.ID)
		}
	}
	return fn(g, AnyPlayerEachAnswer{
		Source: st.source, Controller: offer.Controller, Unsaved: unsaved,
		Carry: copyUUIDs(st.carry), Item: st.asked.StackItem(),
	})
}

// settleAnyPlayerEachWithoutAsking is the offer when nobody is asked at
// all: the consequence runs at once on every object it was given.
func (g *Game) settleAnyPlayerEachWithoutAsking(p AnyPlayerPaysForEach, refs []ObjectRef) error {
	fn, ok := lookupAnyPlayerEachFunc(p.Then.key)
	if !ok {
		return fmt.Errorf("%w: per-object any-player consequence %q", ErrUnknownEffectKey, p.Then.key)
	}
	offer := &AnyPlayerOffer{Controller: p.Controller, EachLife: p.EachLife, Objects: cloneObjectRefs(refs), Noun: p.Noun}
	st := eachOfferState{key: p.Then.key, fn: fn, source: p.Source, carry: copyUUIDs(p.Carry), asked: copyAskedItem(p.Asked)}
	return g.settleAnyPlayerEachLocked(st, offer)
}
