package game

import (
	"errors"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// conjure.go — "conjure a card named <name> into <zone>" (ADR 0099
// amendment 2026-11-24, #5948).
//
// Conjure is an Alchemy keyword action. The Comprehensive Rules do not
// define it; the cards' reminder text does: "(Treat conjured cards like
// cards.)" A conjured card is a new CARD — not a token, not a copy —
// with the printed characteristics of the named card, owned by the
// player who conjured it, put into the named zone. It came from no zone
// and was in no deck.
//
// # Where the card comes from
//
// The engine holds no card database. The catalog files a printed
// template for every card some conjure names (RegisterConjurable, called
// from effects.RegisterConjurable at init), and a conjure builds its new
// object from that template by NAME, the way the card says it. A name no
// template answers is ErrNotConjurable: the card file is wrong, and an
// effect that silently did nothing would be weaker than printed.
//
// The registry is a package-level map, not a field of Game, so nothing
// here is reachable from a game (the closure ratchet,
// closure_fields.txt), and a conjured object is ordinary Card data from
// the moment it exists: a snapshot writes it as it writes any card.
//
// # The zones
//
//   - Hand: pushed onto the owner's hand, known to the owner only
//     (revealed to everyone when the card says "and reveal it").
//   - Library: pushed into the owner's library, known to nobody. Every
//     printed library conjure is followed by "then shuffle", which the
//     caller does once for the whole instruction.
//   - Battlefield: entered through the ordinary CR 614 entry pipeline,
//     staged the way a created token is (Game.enteringTokens), so "enters
//     tapped" effects, entry counters and the enters triggers all see it.
//     Its arrival is announced as a move from no zone (OldZone empty), as
//     a spawned card's is, because a conjured card is not a token and is
//     not "created".
//
// Every conjure emits EventConjured, one per card, before the card
// lands, so the log reads "X conjured Mox Jet into their hand" and then
// whatever the arrival set off. A hand or library conjure names the zone
// and not the card in the public log, as a spawn does (protocol
// projectEvent): the name is hidden information about a hidden zone.

// ErrNotConjurable is returned for a conjure of a name no catalog
// template answers.
var ErrNotConjurable = errors.New("game: no conjurable card has that name")

var (
	conjurableMu sync.RWMutex
	conjurables  = map[string]Card{}
)

// RegisterConjurable files the printed template `c` under c.Name, for
// conjure to build new card objects from. Called from the catalog's
// init; a second template under one name panics, as a duplicate catalog
// registration does.
func RegisterConjurable(c Card) {
	if c.Name == "" {
		panic("game: RegisterConjurable needs a name")
	}
	conjurableMu.Lock()
	defer conjurableMu.Unlock()
	if _, dup := conjurables[c.Name]; dup {
		panic("game: conjurable " + c.Name + " registered twice")
	}
	conjurables[c.Name] = c
}

// ConjurableTemplate is the printed template filed under `name`.
func ConjurableTemplate(name string) (Card, bool) {
	conjurableMu.RLock()
	defer conjurableMu.RUnlock()
	c, ok := conjurables[name]
	return c, ok
}

// ConjurableNames lists every registered name, sorted.
func ConjurableNames() []string {
	conjurableMu.RLock()
	defer conjurableMu.RUnlock()
	out := make([]string, 0, len(conjurables))
	for n := range conjurables {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// ConjureSpec is one conjure instruction.
type ConjureSpec struct {
	// Player conjures and owns the new cards.
	Player uuid.UUID
	// Source is the object whose ability is resolving: the log names
	// it, and a battlefield entry carries it.
	Source uuid.UUID
	// Names are the cards conjured, one card per entry, in order
	// ("conjure four cards named Lightning Bolt" is four entries; "the
	// Power Nine" is nine).
	Names []string
	// Zone is ZoneHand, ZoneLibrary or ZoneBattlefield.
	Zone ZoneKind
	// Reveal reveals each card conjured into a hand ("… into your hand
	// and reveal it").
	Reveal bool
	// Then runs once every card has landed, with the IDs of those that
	// did (a battlefield entry may pause on a prompt, so this is how a
	// caller continues). Zero for none.
	Then KeyedTail
}

// ConjureForEffect performs a conjure. It returns the IDs of the cards
// that landed by the time it returned (a battlefield entry that paused
// lands later, and reaches Then instead).
//
// Caller must hold g.mu in write mode.
func (g *Game) ConjureForEffect(spec ConjureSpec) ([]uuid.UUID, error) {
	p := g.playerByIDLocked(spec.Player)
	if p == nil {
		return nil, ErrPlayerNotFound
	}
	templates := make([]Card, 0, len(spec.Names))
	for _, n := range spec.Names {
		t, ok := ConjurableTemplate(n)
		if !ok {
			return nil, ErrNotConjurable
		}
		templates = append(templates, t)
	}
	switch spec.Zone {
	case ZoneHand, ZoneLibrary:
		return g.conjureIntoOwnedZoneLocked(p, spec, templates)
	case ZoneBattlefield:
		return g.conjureOntoBattlefieldLocked(p, spec, templates)
	}
	return nil, ErrInvalidParam
}

// ConjureRandomNameForEffect is "conjure a card named A, B, or C at
// random": one of `names`, drawn from the keyed pick stream (ADR 0054),
// so a restore replays it.
//
// Caller must hold g.mu in write mode.
func (g *Game) ConjureRandomNameForEffect(player, source uuid.UUID, names []string) string {
	if len(names) == 0 {
		return ""
	}
	r := g.randForLocked(rngStream{kind: rngStreamPick, player: player, source: source})
	return names[r.IntN(len(names))]
}

// mintConjuredLocked builds one new card object from a template: a
// fresh instance, owned and controlled by `owner`, with nothing of any
// earlier object on it.
func (g *Game) mintConjuredLocked(t Card, owner uuid.UUID) Card {
	c := t
	c.InstanceID = uuid.New()
	g.noteCreatedSourceLocked(c.InstanceID)
	c.Owner = owner
	c.Controller = owner
	c.Counters = nil
	c.CounterStampedAt = nil
	c.Stickers = nil
	c.LostLastCounter = false
	c.Tapped = false
	c.KnownBy = nil
	c.Conjured = true
	return c
}

func (g *Game) emitConjuredLocked(spec ConjureSpec, name string) {
	g.EmitEvent(Event{
		Kind:    EventConjured,
		Actor:   spec.Player,
		Source:  spec.Source,
		Target:  spec.Player,
		NewZone: spec.Zone,
		Label:   name,
		Amount:  1,
	})
}

func (g *Game) conjureIntoOwnedZoneLocked(p *Player, spec ConjureSpec, templates []Card) ([]uuid.UUID, error) {
	dest := p.Hand
	if spec.Zone == ZoneLibrary {
		dest = p.Library
	}
	if dest == nil {
		return nil, ErrPlayerNotFound
	}
	ids := make([]uuid.UUID, 0, len(templates))
	for _, t := range templates {
		c := g.mintConjuredLocked(t, p.ID)
		if spec.Zone == ZoneHand {
			c.AddKnower(p.ID)
		}
		g.emitConjuredLocked(spec, t.Name)
		dest.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	if spec.Zone == ZoneHand && spec.Reveal {
		g.RevealForEffect(RevealSpec{Player: p.ID, Source: spec.Source, Cards: ids, Reason: "conjured"})
	}
	// A card in a hand or library can change what a static reads
	// (hand-size counts, "cards in your library"), and nothing else
	// announced the arrival.
	g.layerVersion.Add(1)
	return ids, g.runKeyedTailLocked(spec.Then, ids, 0)
}

func (g *Game) conjureOntoBattlefieldLocked(p *Player, spec ConjureSpec, templates []Card) ([]uuid.UUID, error) {
	steps := make([]tokenEntryStep, 0, len(templates))
	for _, t := range templates {
		c := g.mintConjuredLocked(t, p.ID)
		for _, seat := range g.Seats {
			if seat != nil {
				c.AddKnower(seat.ID)
			}
		}
		g.emitConjuredLocked(spec, t.Name)
		g.enteringTokens = append(g.enteringTokens, c)
		steps = append(steps, tokenEntryStep{ID: c.InstanceID})
	}
	var landed []uuid.UUID
	err := g.enterStagedTokensLocked(&tokenEntryRun{
		Rest:     steps,
		Source:   spec.Source,
		Then:     snapshotKeyedTail(spec.Then),
		Conjured: true,
		report:   &landed,
	})
	return landed, err
}

// GainKeywordPerpetuallyForEffect is "<card> perpetually gains
// <keyword>" (ADR 0099 amendment 2026-11-24): the keyword is added to
// the CARD's own perpetual list, which every zone change keeps and the
// printed baseline reads (Card.PerpetualKeywords). The card may be in
// any zone. A keyword the card already has perpetually is not added
// twice. A card that is gone (ErrCardNotFound) is not an error for the
// caller to handle: the ability simply has nothing to change, so it
// returns nil.
//
// Caller must hold g.mu in write mode.
func (g *Game) GainKeywordPerpetuallyForEffect(cardID uuid.UUID, keyword string) error {
	kw, ok := CanonicalKeyword(keyword)
	if !ok {
		return ErrInvalidParam
	}
	c := g.findCardByIDLocked(cardID)
	if c == nil {
		return nil
	}
	for _, have := range c.PerpetualKeywords {
		if have == kw {
			return nil
		}
	}
	// A fresh slice: the printed cache and a clone may share the old one.
	c.PerpetualKeywords = append(append([]string(nil), c.PerpetualKeywords...), kw)
	c.stampPrinted()
	g.layerVersion.Add(1)
	g.recomputeLayersLocked()
	return nil
}
