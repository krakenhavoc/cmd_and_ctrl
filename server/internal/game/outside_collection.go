package game

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// outside_collection.go — "a card you own from outside the game" for a
// player whose deck seated no sideboard (ADR 0198).
//
// ADR 0191 made a deck's sideboard that player's outside-the-game pool
// and rejected a free pick from the card database. The owner reversed
// half of that on 2026-10-15: a deck WITH a sideboard still chooses from
// it, and a deck WITHOUT one (plain Commander, where CR 903.5e allows
// none) chooses from the whole card database — the casual reading of
// Scryfall's ruling, "a card you choose from outside the game comes from
// your personal collection".
//
// The pick is one prompt kind, PendingChoiceOutsideCollection, answered
// with card NAMES. The answer is checked against the effect's filter and
// set rule, and each named card is then MINTED: a real card, owned by
// the chooser, built the way a deck import builds one (so its catalog
// hooks apply), and put into the chooser's outside-the-game container,
// Player.Sideboard. From there the effect's continuation runs exactly as
// it does for a sideboard pick — the same reveal, the same move into the
// hand, the same events — so nothing downstream knows or cares which pool
// the card came from.
//
// Three things keep it honest:
//
//   - the filter runs on the minted card's printed characteristics, on
//     submit, with the same Pred / Validate pair a library search uses
//     (the frame IS a library search's frame, searchResumeFrame, so the
//     closure census gains no route);
//   - the human picker searches the database through
//     SearchOutsideCollection, which applies the same filter, so it only
//     ever shows a legal card;
//   - the enumerator and the bots read a short, data-only suggestion list
//     stamped on the prompt (CollectionSuggestions), so a seat that
//     cannot type a name still has answers, and a mandatory pick always
//     has one.

// PendingChoiceOutsideCollection is "choose a card you own from outside
// the game" when the chooser's pool is the card database (ADR 0198). It
// is answered with `{card_names: [...]}`; the empty list declines when
// the prompt's floor is zero.
const PendingChoiceOutsideCollection PendingChoiceKind = "outside_collection"

// collectionSuggestionCap is how many suggestions a collection prompt
// carries. Small: they are a convenience for the client and the answers
// the enumerator offers a bot, not a vocabulary — the database is the
// vocabulary, and the search reaches all of it.
const collectionSuggestionCap = 8

// collectionSearchCap bounds one search of the database.
const collectionSearchCap = 40

// OutsideCollectionSource is the card database seen as a player's
// collection. internal/game cannot import internal/cards or
// internal/deck (EffectResolver explains the cycle), so main.go installs
// an implementation built on the loaded index (deck.NewCollection).
//
// Every Card it returns is a TEMPLATE in deck.ToGameCard's shape, with
// no owner; its instance ID means nothing. The game mints from it.
type OutsideCollectionSource interface {
	// Search returns up to `limit` playable cards whose name matches
	// `q`, best match first (exact, then prefix, then contains).
	Search(q string, limit int) []Card
	// Named returns the playable card with this exact name,
	// case-insensitive, or false.
	Named(name string) (Card, bool)
	// Suggest walks the collection's suggestion order — cards the
	// engine automates first, then the most reprinted — and calls fn
	// with each card until fn returns false.
	Suggest(fn func(Card) bool)
}

// OutsideCollection is the installed collection, or nil when the server
// runs with no card index. Nil means a player with no sideboard has
// nothing outside the game, which is ADR 0191's behaviour and the state
// of every game-package test that does not install one.
var OutsideCollection OutsideCollectionSource

// ErrNotACollectionPick is returned by SearchOutsideCollection for a
// choice that is not an open collection pick owed by the caller.
var ErrNotACollectionPick = errors.New("game: not an open outside-the-game pick for this player")

// HasSideboardPool reports whether `player`'s outside-the-game pool is
// their deck's sideboard (ADR 0191) rather than the card database (ADR
// 0198). It is whether their deck SEATED a sideboard, not whether the
// container still holds a card: a sideboard emptied by wishes is still
// the pool, and an empty pool brings nothing in.
//
// Caller must hold g.mu (read or write).
func (g *Game) HasSideboardPool(player uuid.UUID) bool {
	p := g.playerByIDLocked(player)
	return p != nil && p.HasSideboard
}

// outsidePoolOpenLocked is whether `player` may be offered anything from
// outside the game right now: seated, not eliminated, and making their
// own decisions (ADR 0167 §5, CR 723.4).
func (g *Game) outsidePoolOpenLocked(player uuid.UUID) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	return g.ControllerOfPlayerLocked(player) == player
}

// CollectionSuggestionsForEffect is the suggestion list a collection
// pick for `match` would carry: up to collectionSuggestionCap distinct
// card names in the source's suggestion order. Empty when no source is
// installed or nothing in the database matches.
func CollectionSuggestionsForEffect(match func(Card) bool, limit int) []string {
	src := OutsideCollection
	if src == nil || limit <= 0 {
		return nil
	}
	var out []string
	seen := map[string]bool{}
	src.Suggest(func(c Card) bool {
		if match != nil && !match(c) {
			return true
		}
		key := strings.ToLower(c.Name)
		if c.Name == "" || seen[key] {
			return true
		}
		seen[key] = true
		out = append(out, c.Name)
		return len(out) < limit
	})
	return out
}

// queueCollectionPickLocked is QueueOutsidePickForEffect's half for a
// player whose pool is the card database. It reports false, and queues
// nothing, when no collection is installed or nothing in it matches.
func (g *Game) queueCollectionPickLocked(p OutsidePickPrompt) bool {
	if OutsideCollection == nil {
		return false
	}
	max := p.Max
	if max <= 0 {
		max = 1
	}
	min := p.Min
	if min < 0 {
		min = 0
	}
	if min > max {
		min = max
	}
	// Enough suggestions that a Max-2 set rule (two different names) has
	// pairs to offer, and never fewer than the cap.
	limit := collectionSuggestionCap
	if max > limit {
		// Bounded: an "any number" pick (OutsideAnyNumber) does not
		// need the whole database listed, only enough to start from.
		limit = max
		if limit > collectionSearchCap {
			limit = collectionSearchCap
		}
	}
	suggestions := CollectionSuggestionsForEffect(p.Match, limit)
	if len(suggestions) == 0 {
		return false
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:                  PendingChoiceOutsideCollection,
		Chooser:               p.Player,
		FromPlayer:            p.Player,
		Count:                 max,
		Source:                p.Source,
		Reason:                p.Question,
		ChooseMin:             min,
		ChooseMax:             max,
		ChooseDestination:     p.outsideDestination(),
		CollectionSuggestions: suggestions,
		// A library search's frame: Pred is the per-card filter, Validate
		// the set rule, Then the continuation. Reused so the closure
		// census (testdata/closure_fields.txt) gains no route.
		searchResume: &searchResumeFrame{spec: SearchLibrarySpec{
			Player:   p.Player,
			Source:   p.Source,
			Pred:     p.Match,
			Validate: p.Validate,
			Then:     p.Then,
			Reason:   p.Question,
		}},
	})
	return true
}

// collectionChoiceLocked finds the open collection pick `choiceID`.
func (g *Game) collectionChoiceLocked(choiceID uuid.UUID) (int, *PendingChoice) {
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID && c.Kind == PendingChoiceOutsideCollection {
			return i, c
		}
	}
	return -1, nil
}

// collectionTemplatesLocked resolves and checks an answer to a
// collection pick: the count is within the bounds, every name is a card
// in the database that satisfies the filter, and the named set passes
// the set rule. The templates come back in submitted order.
//
// The one copy of "would this answer be accepted": the resolver runs it
// before it dequeues and the enumerator runs it (CollectionPickLegalLocked)
// before it offers an answer.
func collectionTemplatesLocked(choice *PendingChoice, names []string) ([]Card, error) {
	if choice == nil || choice.searchResume == nil {
		return nil, ErrInvalidParam
	}
	if len(names) < choice.ChooseMin || len(names) > choice.ChooseMax {
		return nil, ErrInvalidParam
	}
	if len(names) == 0 {
		return nil, nil
	}
	src := OutsideCollection
	if src == nil {
		return nil, ErrInvalidParam
	}
	spec := choice.searchResume.spec
	out := make([]Card, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || len(n) > MaxChosenNameLen {
			return nil, ErrInvalidParam
		}
		c, ok := src.Named(n)
		if !ok {
			return nil, ErrCardNotFound
		}
		if spec.Pred != nil && !spec.Pred(c) {
			return nil, ErrInvalidParam
		}
		out = append(out, c)
	}
	if spec.Validate != nil && !spec.Validate(out) {
		return nil, ErrInvalidParam
	}
	return out, nil
}

// CollectionPickLegalLocked reports whether ResolveOutsideCollection
// would accept `names` for `choice` right now. The enumerator's window
// onto the frame's filter and set rule, the way SearchPickLegalLocked is
// a search's.
//
// Caller must hold g's lock (read or write).
func (g *Game) CollectionPickLegalLocked(choice *PendingChoice, names []string) bool {
	if choice == nil || choice.Kind != PendingChoiceOutsideCollection {
		return false
	}
	_, err := collectionTemplatesLocked(choice, names)
	return err == nil
}

// CollectionCardsLocked resolves a collection prompt's suggestion names
// to their templates, for the view (which shows them as cards) and the
// bots (which value them). A name the database no longer has is skipped.
//
// Caller must hold g's lock (read or write).
func CollectionCardsLocked(names []string) []Card {
	src := OutsideCollection
	if src == nil {
		return nil
	}
	out := make([]Card, 0, len(names))
	for _, n := range names {
		if c, ok := src.Named(n); ok {
			out = append(out, c)
		}
	}
	return out
}

// SearchOutsideCollection is the human picker's search (ADR 0198 §3):
// the cards in the database whose name matches `q` AND that the open
// collection pick `choiceID` would accept one at a time, best match
// first, at most `limit` (capped at collectionSearchCap).
//
// Only the pick's chooser may search it: ErrNotACollectionPick for any
// other caller, any other kind and a closed pick. The set rule is not
// applied — it judges a set, and a search row is one card.
//
// Takes the read lock.
func (g *Game) SearchOutsideCollection(choiceID, player uuid.UUID, q string, limit int) ([]Card, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, choice := g.collectionChoiceLocked(choiceID)
	if choice == nil || choice.Chooser != player || choice.searchResume == nil {
		return nil, ErrNotACollectionPick
	}
	src := OutsideCollection
	q = strings.TrimSpace(q)
	if src == nil || q == "" {
		return nil, nil
	}
	if limit <= 0 || limit > collectionSearchCap {
		limit = collectionSearchCap
	}
	pred := choice.searchResume.spec.Pred
	// Over-fetch: the filter may reject most of a name match ("Burning
	// Wish" against "wish" is mostly creatures and enchantments).
	var out []Card
	for _, c := range src.Search(q, limit*8) {
		if pred != nil && !pred(c) {
			continue
		}
		out = append(out, c)
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

// ResolveOutsideCollection answers a PendingChoiceOutsideCollection.
// `names` are the cards the chooser takes, in order; empty declines when
// the prompt's floor is zero.
//
// Each named card is minted (mintFromCollectionLocked) into the chooser's
// outside-the-game container and the effect's continuation receives the
// new instance IDs, exactly as a sideboard pick's continuation receives
// the IDs of the cards it named. When the pool has closed since the
// prompt opened (the chooser left, or another player now makes their
// decisions) the answer is taken as "nothing": no card is minted.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveOutsideCollection(choiceID, chooserID uuid.UUID, names []string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.collectionChoiceLocked(choiceID)
	if choice == nil {
		if _, ok := g.pendingChoiceIndexLocked(choiceID); ok {
			return ErrInvalidParam
		}
		return ErrPendingChoiceNotFound
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	templates, err := collectionTemplatesLocked(choice, names)
	if err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)

	var ids []uuid.UUID
	if g.outsidePoolOpenLocked(chooserID) {
		for _, t := range templates {
			ids = append(ids, g.mintFromCollectionLocked(chooserID, t))
		}
	}
	if then := choice.searchResume.spec.Then; then != nil {
		if err := then(g, ids); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    chooserID,
				Source:   choice.Source,
				ErrorMsg: err.Error(),
			})
		}
	}
	g.runStateChecksLocked()
	return nil
}

// pendingChoiceIndexLocked finds any open choice by ID.
func (g *Game) pendingChoiceIndexLocked(choiceID uuid.UUID) (int, bool) {
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			return i, true
		}
	}
	return -1, false
}

// mintFromCollectionLocked makes the card a player takes from their
// collection: a fresh instance of `template`, owned and controlled by
// them, known only to them, in their outside-the-game container. It is
// the card a deck import would have made (the template is
// deck.ToGameCard's), so every catalog hook keyed by its oracle ID
// applies. Returns its instance ID.
//
// The container is the sideboard's (ADR 0191), created when the player
// has none; HasSideboard is NOT set, so their pool stays the database.
// A card the effect then leaves outside (Extrapolate the Impossible's
// unchosen card) simply stays there, outside the game, as it would on a
// table.
//
// Caller must hold g.mu.
func (g *Game) mintFromCollectionLocked(player uuid.UUID, template Card) uuid.UUID {
	p := g.playerByIDLocked(player)
	if p == nil {
		return uuid.Nil
	}
	c := template
	c.InstanceID = uuid.New()
	g.noteCreatedSourceLocked(c.InstanceID)
	c.Owner = player
	c.Controller = player
	c.IsSideboard = false
	c.IsCommander = false
	c.IsCompanion = false
	c.Counters = nil
	c.CounterStampedAt = nil
	c.LostLastCounter = false
	c.Tapped = false
	c.ClearKnown()
	c.AddKnower(player)
	if p.Sideboard == nil {
		p.Sideboard = newZone(ZoneOutside, p.ID)
	}
	p.Sideboard.PushTop(c)
	return c.InstanceID
}
