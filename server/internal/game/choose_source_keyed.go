package game

import (
	"fmt"

	"github.com/google/uuid"
)

// choose_source_keyed.go — #4253 slice 10c: a choose_source prompt ("choose
// a source of damage", CR 609.7a) whose consequence is a KEY rather than a
// closure, so a table waiting on the answer is a restore point.
//
// The prompt was a card-set pick (PendingChoiceChooseSource) that held its
// continuation as a Go closure on the prompt's chooseCardsResume frame. It
// now rides the keyed choose_cards machinery (choose_cards_keyed.go): a
// continuation registered under an on-disk key (RegisterChooseSourceThen,
// ledgered "choose <key>"), handed the source as it was pinned and the plain
// data the card put on the prompt (Carry, Word, Nums, the asking item).
//
// Most cards that choose a source do one thing with it: make a shield, a
// redirection or a multiplier. The three shields are plain data (a
// NextDamageShield, a DamageShield, a DamageRedirection) the card finished
// describing before it asked, and the engine's own continuation completes
// one with the pinned source and registers it (SourceShield, below). A card
// that does something else with the source registers a continuation of its
// own (Desperate Gambit's coin flip).

// ChosenSource is what a keyed choose_source continuation is handed.
type ChosenSource struct {
	// Chooser answered; Source is the card that asked.
	Chooser, Source uuid.UUID
	// Ref is the source as it is NOW (CR 609.7a: "chosen when the effect is
	// created"), in Zone. The zero ref when nothing could be chosen: no
	// legal source, or a chooser who has left.
	Ref  ObjectRef
	Zone ZoneKind
	// Carry, Word and Nums are the plain data the queuing card put on the
	// prompt; Item is the stack item that asked, rebuilt from the data the
	// prompt carried, or nil when the card put none there.
	Carry []uuid.UUID
	Word  string
	Nums  []int
	Item  *StackItem

	shield *SourceShield
}

// ChooseSourceFunc is a registered choose_source continuation. It runs with
// g.mu held, may queue prompts of its own, and must capture nothing.
type ChooseSourceFunc func(g *Game, r ChosenSource) error

// ChooseSourceThen names a registered choose_source continuation. The key
// is unexported, so the only way to hold one is RegisterChooseSourceThen.
type ChooseSourceThen struct{ then ChooseCardsThen }

// Key is the continuation's on-disk key.
func (t ChooseSourceThen) Key() string { return t.then.key }

// IsZero reports whether no continuation is named.
func (t ChooseSourceThen) IsZero() bool { return t.then.key == "" }

// chooseSourceFns are the registered continuations' bodies, by key, for a
// prompt that could not be asked. Guarded by effectRegistryMu.
var chooseSourceFns = map[string]ChooseSourceFunc{}

// RegisterChooseSourceThen registers a keyed choose_source continuation.
// Call it once, from a package-level var. It shares the choose_cards
// namespace and ledger (a "choose <key>" line), and the key is an on-disk
// identity: never renamed, never reused.
func RegisterChooseSourceThen(key string, fn ChooseSourceFunc) ChooseSourceThen {
	if fn == nil {
		panic(fmt.Sprintf("game: choose-source continuation %q has no function", key))
	}
	then := RegisterChooseCardsThen(key, func(g *Game, r CardsChosen) error {
		return g.finishChosenSource(r, fn)
	})
	effectRegistryMu.Lock()
	chooseSourceFns[key] = fn
	effectRegistryMu.Unlock()
	return ChooseSourceThen{then: then}
}

// finishChosenSource is the keyed pick's continuation: it narrates the
// answer, pins the source as it is now and hands it on. An answer that is
// not a source any more (it has left every zone a source is read in) is
// the zero ref, as a prompt nobody answered is.
func (g *Game) finishChosenSource(r CardsChosen, fn ChooseSourceFunc) error {
	g.emitCardsChosenLocked(r.Chooser, r.Source, r.Picked)
	cs := ChosenSource{
		Chooser: r.Chooser, Source: r.Source,
		Carry: r.Carry, Word: r.Word, Nums: r.Nums, Item: r.Asked,
		shield: r.shield,
	}
	if len(r.Picked) == 1 {
		if ref, zone, ok := g.DamageSourceRefLocked(r.Picked[0]); ok {
			cs.Ref, cs.Zone = ref, zone
		}
	}
	return fn(g, cs)
}

// SourceShield is a choose_source prompt's consequence when it is the
// engine's own: the shield, redirection or multiplier the card described,
// completed with the source once it is pinned. Exactly one field is set.
// Plain data, written with the prompt.
type SourceShield struct {
	// Next is "the next time a source of your choice would deal damage ...
	// prevent that damage" (ADR 0107 §6).
	Next *NextDamageShield
	// Prevent is a shield that lasts the turn (PreventDamageFromSource).
	Prevent *DamageShield
	// Redirect is a damage redirection (RedirectDamage).
	Redirect *DamageRedirection
}

// Register completes the shield with the source (the zero ref is "any
// source matching the shield's queries") and registers it for the rest of
// the turn.
//
// Caller must hold g.mu (write).
func (s SourceShield) Register(g *Game, ref ObjectRef, zone ZoneKind) {
	switch {
	case s.Next != nil:
		n := *s.Next
		n.Source, n.SourceZone = ref, zone
		g.PreventNextDamageFromSourceForEffect(n)
	case s.Prevent != nil:
		p := *s.Prevent
		p.Source, p.SourceZone = ref, zone
		g.PreventDamageFromSourceThisTurnForEffect(p)
	case s.Redirect != nil:
		r := *s.Redirect
		r.Source, r.SourceZone = ref, zone
		g.RedirectDamageThisTurnForEffect(r)
	}
}

// chooseSourceShieldThen is the engine's continuation for a SourceShield:
// no source chosen is no shield.
var chooseSourceShieldThen = RegisterChooseSourceThen("choose-source/shield", func(g *Game, r ChosenSource) error {
	if r.Ref.ID == uuid.Nil || r.shield == nil {
		return nil
	}
	r.shield.Register(g, r.Ref, r.Zone)
	return nil
})

// cloneSourceShield gives a prompt its own copy. The shields hold slices;
// they are copied by value on use (every consumer copies the struct and
// the registrations clone their slices), so a shallow copy of the struct is
// what an undo clone needs.
func cloneSourceShield(s *SourceShield) *SourceShield {
	if s == nil {
		return nil
	}
	out := &SourceShield{}
	if s.Next != nil {
		n := *s.Next
		n.Queries = clonePermanentQueries(n.Queries)
		n.ProtectTypes = copyStrings(n.ProtectTypes)
		out.Next = &n
	}
	if s.Prevent != nil {
		p := *s.Prevent
		p.Queries = clonePermanentQueries(p.Queries)
		p.ProtectTypes = copyStrings(p.ProtectTypes)
		p.ProtectPermanents = copyUUIDs(p.ProtectPermanents)
		p.Filter.Except = clonePermanentQueries(p.Filter.Except)
		p.Filter.ExceptObjects = cloneObjectRefs(p.Filter.ExceptObjects)
		p.Filter.NoCounters = copyStrings(p.Filter.NoCounters)
		p.Recipients = p.Recipients.clone()
		out.Prevent = &p
	}
	if s.Redirect != nil {
		r := *s.Redirect
		r.Queries = clonePermanentQueries(r.Queries)
		r.ProtectTypes = copyStrings(r.ProtectTypes)
		r.SourceFilter.Except = clonePermanentQueries(r.SourceFilter.Except)
		r.SourceFilter.ExceptObjects = cloneObjectRefs(r.SourceFilter.ExceptObjects)
		r.SourceFilter.NoCounters = copyStrings(r.SourceFilter.NoCounters)
		r.ProtectSet = r.ProtectSet.clone()
		out.Redirect = &r
	}
	return out
}

// sourceShieldSnapshot is a SourceShield on disk. The shields' bodies are
// registered keys (BodyRef hides its key from JSON), so each wire form
// carries the key beside the struct.
type sourceShieldSnapshot struct {
	Next     *nextShieldWire   `json:"next,omitempty"`
	Prevent  *damageShieldWire `json:"prevent,omitempty"`
	Redirect *redirectionWire  `json:"redirect,omitempty"`
}

type nextShieldWire struct {
	NextDamageShield
	ThenKey string `json:"thenKey,omitempty"`
}

type damageShieldWire struct {
	DamageShield
	ThenKey string `json:"thenKey,omitempty"`
}

type redirectionWire struct {
	DamageRedirection
	ThenKey string `json:"thenKey,omitempty"`
}

func snapshotSourceShield(s *SourceShield) *sourceShieldSnapshot {
	if s == nil {
		return nil
	}
	s = cloneSourceShield(s)
	out := &sourceShieldSnapshot{}
	if s.Next != nil {
		out.Next = &nextShieldWire{NextDamageShield: *s.Next, ThenKey: s.Next.Then.key}
	}
	if s.Prevent != nil {
		out.Prevent = &damageShieldWire{DamageShield: *s.Prevent, ThenKey: s.Prevent.Then.key}
	}
	if s.Redirect != nil {
		out.Redirect = &redirectionWire{DamageRedirection: *s.Redirect, ThenKey: s.Redirect.Then.key}
	}
	return out
}

func restoreSourceShield(s *sourceShieldSnapshot) *SourceShield {
	if s == nil {
		return nil
	}
	out := &SourceShield{}
	if s.Next != nil {
		n := s.Next.NextDamageShield
		n.Then = BodyRef{key: s.Next.ThenKey}
		out.Next = &n
	}
	if s.Prevent != nil {
		p := s.Prevent.DamageShield
		p.Then = BodyRef{key: s.Prevent.ThenKey}
		out.Prevent = &p
	}
	if s.Redirect != nil {
		r := s.Redirect.DamageRedirection
		r.Then = BodyRef{key: s.Redirect.ThenKey}
		out.Redirect = &r
	}
	return out
}

// unknownSourceShieldKeys is checkEffectKeys' line for a shield whose
// follow-up body this binary does not register: restored without it, the
// prevented damage would do nothing more.
func unknownSourceShieldKeys(s *sourceShieldSnapshot) []string {
	if s == nil {
		return nil
	}
	var out []string
	for _, key := range []string{s.nextKey(), s.preventKey(), s.redirectKey()} {
		if key != "" && !KnownEffectBody(key) {
			out = append(out, "stack-item body "+key)
		}
	}
	return out
}

func (s *sourceShieldSnapshot) nextKey() string {
	if s.Next == nil {
		return ""
	}
	return s.Next.ThenKey
}

func (s *sourceShieldSnapshot) preventKey() string {
	if s.Prevent == nil {
		return ""
	}
	return s.Prevent.ThenKey
}

func (s *sourceShieldSnapshot) redirectKey() string {
	if s.Redirect == nil {
		return ""
	}
	return s.Redirect.ThenKey
}
