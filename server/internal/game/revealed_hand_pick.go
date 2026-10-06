package game

import (
	"fmt"
	"slices"
	"sort"

	"github.com/google/uuid"
)

// revealed_hand_pick.go — the variants of ADR 0116's revealed-hand pick
// (#2115, ADR 0116's 2026-10-05 amendment).
//
// "Target opponent reveals their hand. You choose a nonland card from
// it. That player discards that card." is QueueDiscardFromRevealedHand
// and its discard_from_hand prompt. The printed variants of the same
// sentence differ in three ways, and a card may combine them:
//
//   - WHERE the card goes. Appetite for Brains, Aggressive Negotiations
//     and Agonizing Remorse exile it. Exiling a card from a hand is not
//     discarding it (CR 701.9a: "To discard a card, move it from its
//     owner's hand to that player's graveyard"), so madness (CR 702.35a,
//     "If a player would discard this card") and "whenever a player
//     discards" never see it. RevealedHandDiscard.Destination.
//   - WHETHER a card must be chosen. "You may choose a nonland card
//     from it" (Nightsnare, Binding Negotiation, Reckoner Shakedown,
//     Extract the Truth, Revealing Eye) lets the chooser choose
//     nothing. RevealedHandDiscard.Optional.
//   - WHAT happens next, once it is known what was chosen: "If you do
//     … If you don't …" (Nightsnare), a clause that reads the chosen
//     card (Talara's Bane: "You gain life equal to that creature card's
//     toughness"), or anything printed after the move that must wait
//     for it (Tourach's Canticle: "then discards a card at random").
//     RevealedHandDiscard.Then, and — for a number read off the chosen
//     card — RevealedHandDiscard.Measure, which reads every candidate
//     while the spell is still resolving and hands the continuation
//     the chosen cards' numbers.
//
// Agonizing Remorse adds a fourth, "or a card from their graveyard":
// RevealedHandDiscard.FromGraveyard.
//
// # A restore point, still
//
// ADR 0116 kept the plain pick a restore point by giving it no
// continuation. A variant needs one, so the continuation is KEYED, the
// way ADR 0041 phase 3 keyed delayed triggers (effect_bodies.go): a
// package-level function registered once at init under "<area>/<name>",
// with the prompt carrying only the key. Restore re-derives the
// function from the key and refuses a key this binary does not have
// (ErrUnknownEffectKey). Keys are on-disk identities: never renamed,
// never reused. testdata/effect_keys.txt lists them as "pick <key>".
//
// # Its own kind
//
// A variant is PendingChoiceRevealedHandPick, not a discard_from_hand
// with more fields. A binary from before this change does not know the
// kind, so it REFUSES a restore point holding one (ADR 0115 §8) and the
// file is kept for the roll-forward. Had the variant reused the old
// kind, that binary would have dropped the new fields and restored an
// exile as a discard, or an optional pick as a mandatory one.

// PendingChoiceRevealedHandPick is a revealed-hand pick with a variant:
// another destination, an optional floor, graveyard candidates or a
// keyed continuation. Answered with {card_ids: []string}, empty when
// the pick is optional and the chooser chooses nothing.
const PendingChoiceRevealedHandPick PendingChoiceKind = "revealed_hand_pick"

// PickDestination is where a revealed-hand pick sends the chosen card.
type PickDestination string

const (
	// PickDiscard discards it (CR 701.9a), through the one discard
	// path, so madness and discard triggers apply.
	PickDiscard PickDestination = "discard"
	// PickExile exiles it. Not a discard: no madness, no discard
	// triggers (CR 701.9a).
	PickExile PickDestination = "exile"
)

// knownPickDestination reports whether this binary can move a card to
// d. The zero value is a discard.
func knownPickDestination(d PickDestination) bool {
	return d == "" || d == PickDiscard || d == PickExile
}

// RevealedPick is what a continuation is handed.
type RevealedPick struct {
	// Chooser is the player who chose: the controller of the effect
	// that asked ("you"), or whoever CR 800.4g handed the prompt to.
	// FromPlayer revealed their hand. Source is the card that asked.
	Chooser, FromPlayer, Source uuid.UUID
	// Chosen is the cards chosen, as they were when chosen: value
	// copies read before they moved, which is their last-known
	// information once they have left (CR 608.2h). Empty when nothing
	// was chosen, whether the chooser chose nothing or there was
	// nothing to choose. Read-only: their maps are shared with the live
	// cards.
	Chosen []Card
	// Measures is RevealedHandDiscard.Measure's reading of each chosen
	// card, in Chosen order, taken as the pick was raised. Nil when the
	// card measures nothing.
	Measures []int
	// Done moves the chosen cards. It is set only for a continuation
	// registered with RegisterRevealedPickFirst, which must call it
	// exactly once, from wherever it finishes (inline, or from the Then
	// of something it paused on). For every other continuation it is
	// nil, because the cards have already moved.
	Done func(g *Game) error
}

// ChosenIDs is the chosen cards' instance IDs, in the order chosen.
func (r RevealedPick) ChosenIDs() []uuid.UUID {
	out := make([]uuid.UUID, len(r.Chosen))
	for i, c := range r.Chosen {
		out[i] = c.InstanceID
	}
	return out
}

// RevealedPickFunc is a registered continuation. It runs with g.mu
// held, may queue prompts of its own, and must capture nothing: it is a
// package-level function, and everything it reads is on the RevealedPick
// or the game.
type RevealedPickFunc func(g *Game, r RevealedPick) error

// RevealedPickThen names a registered continuation. The key is
// unexported, so the only way to hold one is RegisterRevealedPickThen
// or RegisterRevealedPickFirst, which keeps a func literal out of a
// prompt.
type RevealedPickThen struct{ key string }

// Key is the continuation's on-disk key.
func (r RevealedPickThen) Key() string { return r.key }

type revealedPickEntry struct {
	fn RevealedPickFunc
	// first: the continuation runs before the chosen cards move, and
	// moves them itself through RevealedPick.Done.
	first bool
}

// revealedPickThens is guarded by effectRegistryMu, with the other
// effect keys, because the keys share one namespace and one ledger.
var revealedPickThens = map[string]revealedPickEntry{}

// RegisterRevealedPickThen registers a continuation that runs after the
// chosen cards have moved: "If you do, that player discards that card,
// then draws a card", "If you don't, that player discards two cards",
// "then discards a card at random". Call it once, from a package-level
// var. Panics on a bad or duplicate key, as DelayedBody does.
func RegisterRevealedPickThen(key string, fn RevealedPickFunc) RevealedPickThen {
	return registerRevealedPick(key, fn, false)
}

// RegisterRevealedPickFirst registers a continuation printed BEFORE the
// move: Talara's Bane's "You gain life equal to that creature card's
// toughness, then that player discards that card" (CR 608.2c). It is
// handed RevealedPick.Done and must call it exactly once.
func RegisterRevealedPickFirst(key string, fn RevealedPickFunc) RevealedPickThen {
	return registerRevealedPick(key, fn, true)
}

func registerRevealedPick(key string, fn RevealedPickFunc, first bool) RevealedPickThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("revealed-hand pick continuation", key)
	if fn == nil {
		panic(fmt.Sprintf("game: revealed-hand pick continuation %q has no function", key))
	}
	revealedPickThens[key] = revealedPickEntry{fn: fn, first: first}
	return RevealedPickThen{key: key}
}

func lookupRevealedPick(key string) (revealedPickEntry, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	e, ok := revealedPickThens[key]
	return e, ok
}

// KnownRevealedPickThen reports whether this binary has a continuation
// registered under key.
func KnownRevealedPickThen(key string) bool { _, ok := lookupRevealedPick(key); return ok }

// registeredRevealedPickKeys lists the ledger lines, "pick <key>".
// Caller holds effectRegistryMu.
func registeredRevealedPickKeysLocked() []string {
	out := make([]string, 0, len(revealedPickThens))
	for k := range revealedPickThens {
		out = append(out, "pick "+k)
	}
	sort.Strings(out)
	return out
}

// RevealedHandPickForEffect is the revealed-hand pick with a variant
// (see the file comment). In order:
//
//  1. FromPlayer reveals their hand to every player (CR 701.20a), as
//     the plain pick does. An empty hand reveals nothing.
//  2. The candidates are fixed: the hand's cards that pass Filter, in
//     hand order, then — with FromGraveyard — every card in
//     FromPlayer's graveyard, in graveyard order (CR 608.2c, as ADR
//     0116 §3).
//  3. No candidate: nothing is chosen (CR 609.3), and the continuation
//     runs now, told that nothing was chosen — "If you don't" happens
//     when there was nothing to choose too. Its error is returned.
//  4. Otherwise a revealed_hand_pick prompt is queued for the chooser,
//     with Count capped at the number of candidates. Its answer moves
//     the cards and runs the continuation (resolveRevealedHandPickLocked).
//
// No opponent-protection check is made here (#2178, Tamiyo, Collector
// of Tales). "Can't cause you to discard" stops the discard, not the
// reveal or the choice, and a card such as Talara's Bane still gains
// the life. The discard is checked when it would happen.
//
// Returns the prompt's ID, or uuid.Nil when nothing was queued.
//
// Caller must hold g.mu.
func (g *Game) RevealedHandPickForEffect(d RevealedHandDiscard) (uuid.UUID, error) {
	if !knownPickDestination(d.Destination) {
		return uuid.Nil, fmt.Errorf("%w: revealed-hand pick destination %q", ErrInvalidParam, d.Destination)
	}
	if d.FromGraveyard && d.Destination != PickExile {
		// A card in a graveyard cannot be discarded (CR 701.9a): the one
		// printed graveyard variant, Agonizing Remorse, exiles.
		return uuid.Nil, fmt.Errorf("%w: a revealed-hand pick from the graveyard must exile", ErrInvalidParam)
	}
	if d.Then.key != "" && !KnownRevealedPickThen(d.Then.key) {
		return uuid.Nil, fmt.Errorf("%w: revealed-hand pick continuation %q", ErrUnknownEffectKey, d.Then.key)
	}
	p := g.playerByIDLocked(d.FromPlayer)
	if p == nil {
		// The player has left the game (CR 800.4a): there is no hand to
		// reveal and nobody to tell.
		return uuid.Nil, nil
	}
	var hand, options []uuid.UUID
	if p.Hand != nil {
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
			if d.Filter == nil || d.Filter(c) {
				options = append(options, c.InstanceID)
			}
		}
	}
	if d.FromGraveyard && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			options = append(options, c.InstanceID)
		}
	}
	reason := d.Reason
	if reason == "" && d.Source != uuid.Nil {
		if card, ok := g.LookupCardForEffect(d.Source); ok {
			reason = card.Name
		}
	}
	g.RevealForEffect(RevealSpec{
		Player: d.FromPlayer,
		Source: d.Source,
		Reason: revealHandReason(reason),
		Cards:  hand,
	})
	count := min(d.Count, len(options))
	if count <= 0 {
		// CR 609.3: nothing to choose. The rest of the card still runs.
		return uuid.Nil, g.settleRevealedPickLocked(revealedPickSettle{
			chooser: d.Chooser, from: d.FromPlayer, source: d.Source,
			destination: d.Destination, then: d.Then.key,
		}, nil, nil)
	}
	if options == nil {
		options = []uuid.UUID{}
	}
	var measures []int
	if d.Measure != nil {
		measures = make([]int, len(options))
		for i, id := range options {
			if card, ok := g.revealedCandidateLocked(p, id, d.FromGraveyard); ok {
				measures[i] = d.Measure(card)
			}
		}
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:              PendingChoiceRevealedHandPick,
		Chooser:           d.Chooser,
		FromPlayer:        d.FromPlayer,
		Count:             count,
		Source:            d.Source,
		Reason:            reason,
		DiscardOptions:    options,
		DiscardLabel:      d.Label,
		PickDestination:   d.Destination,
		PickOptional:      d.Optional,
		PickFromGraveyard: d.FromGraveyard,
		PickThen:          d.Then.key,
		PickMeasures:      measures,
	})
	return id, nil
}

// RevealedPickCandidateLocked reports where a revealed-hand pick's
// candidate is now: in FromPlayer's hand, or — for a pick that offers
// it — their graveyard. ok is false for a card that is in neither, or
// that the prompt never offered. internal/legal and the resolver share
// it, so the bot is offered exactly the picks the resolver accepts.
//
// Caller must hold g.mu (read or write).
func (g *Game) RevealedPickCandidateLocked(c *PendingChoice, id uuid.UUID) (Card, bool) {
	if c == nil || !slices.Contains(c.DiscardOptions, id) {
		return Card{}, false
	}
	return g.revealedCandidateLocked(g.playerByIDLocked(c.FromPlayer), id, c.PickFromGraveyard)
}

// revealedCandidateLocked is the card with that instance ID in p's
// hand or, with graveyard, p's graveyard. Caller must hold g.mu.
func (g *Game) revealedCandidateLocked(p *Player, id uuid.UUID, graveyard bool) (Card, bool) {
	if p == nil {
		return Card{}, false
	}
	if card, ok := findInZone(p.Hand, id); ok {
		return card, true
	}
	if graveyard {
		if card, ok := findInZone(p.Graveyard, id); ok {
			return card, true
		}
	}
	return Card{}, false
}

// findInZone is a value copy of the card with that instance ID in z.
func findInZone(z *Zone, id uuid.UUID) (Card, bool) {
	if z == nil {
		return Card{}, false
	}
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return Card{}, false
}

// resolveRevealedHandPickLocked answers a PendingChoiceRevealedHandPick.
// Refused, with the prompt left open:
//
//   - an empty answer to a pick that is not optional, or an answer of
//     any other size than Count (ErrInvalidParam);
//   - a card the prompt did not offer, or one named twice
//     (ErrInvalidParam, ADR 0116 §5);
//   - a card no longer where it was offered (ErrCardNotFound).
//
// Then the prompt is dequeued, the cards move and the continuation runs
// (settleRevealedPickLocked). An error from either is reported as an
// effect error and the answer still stands — the #544 contract every
// continuation resolver keeps.
//
// Caller must hold g.mu.
func (g *Game) resolveRevealedHandPickLocked(idx int, c *PendingChoice, picks []uuid.UUID) error {
	if g.playerByIDLocked(c.FromPlayer) == nil {
		// FromPlayer left the game with the prompt open (CR 800.4a):
		// their cards went with them, and the question is moot.
		g.dequeueChoiceLocked(idx)
		return nil
	}
	switch {
	case len(picks) == 0 && !c.PickOptional:
		return ErrInvalidParam
	case len(picks) != 0 && len(picks) != c.Count:
		return ErrInvalidParam
	}
	if !discardOptionsAllow(c, picks) {
		return ErrInvalidParam
	}
	chosen := make([]Card, 0, len(picks))
	var measures []int
	for _, id := range picks {
		card, ok := g.RevealedPickCandidateLocked(c, id)
		if !ok {
			return ErrCardNotFound
		}
		chosen = append(chosen, card)
		if c.PickMeasures != nil {
			i := slices.Index(c.DiscardOptions, id)
			if i < 0 || i >= len(c.PickMeasures) {
				// A prompt whose readings do not line up with its
				// candidates was not written by this binary's queue.
				return ErrInvalidParam
			}
			measures = append(measures, c.PickMeasures[i])
		}
	}
	s := revealedPickSettle{
		chooser: c.Chooser, from: c.FromPlayer, source: c.Source,
		destination: c.PickDestination, then: c.PickThen,
	}
	g.dequeueChoiceLocked(idx)
	g.emitChoiceEffectErrorLocked(s.chooser, s.source, g.settleRevealedPickLocked(s, chosen, measures))
	return nil
}

// revealedPickSettle is what a settled pick needs besides the cards:
// plain values copied off the prompt before it was dequeued.
type revealedPickSettle struct {
	chooser, from, source uuid.UUID
	destination           PickDestination
	then                  string
}

// settleRevealedPickLocked moves the chosen cards and runs the
// continuation, in printed order: a RegisterRevealedPickThen
// continuation after the move has finished (a move can pause on a
// replacement's prompt, so it is the move's own continuation), a
// RegisterRevealedPickFirst one before it, holding the move as Done.
//
// Caller must hold g.mu.
func (g *Game) settleRevealedPickLocked(s revealedPickSettle, chosen []Card, measures []int) error {
	var entry revealedPickEntry
	if s.then != "" {
		e, ok := lookupRevealedPick(s.then)
		if !ok {
			return fmt.Errorf("%w: revealed-hand pick continuation %q", ErrUnknownEffectKey, s.then)
		}
		entry = e
	}
	r := RevealedPick{Chooser: s.chooser, FromPlayer: s.from, Source: s.source, Chosen: chosen, Measures: measures}
	ids := r.ChosenIDs()
	move := func(g *Game, after func(g *Game) error) error {
		return g.moveRevealedPickLocked(s, ids, after)
	}
	switch {
	case entry.fn == nil:
		return move(g, nil)
	case entry.first:
		r.Done = func(g *Game) error { return move(g, nil) }
		return entry.fn(g, r)
	default:
		return move(g, func(g *Game) error { return entry.fn(g, r) })
	}
}

// moveRevealedPickLocked sends the chosen cards to the pick's
// destination as one batch, then runs `after` (which may be nil) once
// the batch has finished.
//
// A discard goes through the one discard path, so madness, discard
// triggers and discard replacements all apply. It does not happen when
// an opponent-protection static stops it (#2178): the cause is the
// chooser, the controller of the effect that asked. An exile goes
// through the one exile route, and is not a discard.
//
// Caller must hold g.mu.
func (g *Game) moveRevealedPickLocked(s revealedPickSettle, ids []uuid.UUID, after func(g *Game) error) error {
	next := func(g *Game, _ []uuid.UUID) error {
		if after == nil {
			return nil
		}
		return after(g)
	}
	if len(ids) == 0 {
		return next(g, nil)
	}
	switch s.destination {
	case PickExile:
		return g.ExileCardsThenForEffect(ids, next)
	default:
		if g.opponentEffectBlockedByLocked(s.chooser, s.from, true) {
			return next(g, nil)
		}
		return g.discardCardsLocked(s.from, ids, discardOptions{
			cause:  DiscardCauseEffect,
			source: s.source,
			then:   next,
		})
	}
}
