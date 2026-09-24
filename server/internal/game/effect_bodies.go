package game

import (
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

// effect_bodies.go is ADR 0041 phase 3's tier 2 (Decision P2, #1497):
// a delayed trigger names what it does by KEY rather than holding a
// closure, so a table with one waiting is still a restore point.
//
// A body is a package-level function, registered once at init under a
// key of the form "<area>/<name>" — "flicker/return-exiled-to-owners",
// "warp/exile" — and a delayed trigger carries the key plus the plain
// EffectParams the function reads. When it fires, the stack item
// carries the same two, so it stays data while it waits on the stack
// too. Restore re-derives the function from the key through the
// running binary, and REFUSES a key it does not have
// (ErrUnknownEffectKey): every key a binary writes is one it
// registered, so an unknown one is the rollback case.
//
// Keys are on-disk identities, like token slugs and grant keys:
// append-only. testdata/effect_keys.txt is the ledger, and
// TestEveryPersistedEffectKeyResolves holds it. Rename with
// EffectAlias(old, new); never delete a key.
//
// Event conditions (#663's "when you next cast …") are the same idea
// one slot over: ConditionFunc, registered with DelayedCondition.

// EffectParams is the closed set of values a body or a condition may
// be handed (ADR 0041 P1's allowed types). A body that needs a new
// kind of value gets a new field here — never an interface{}.
type EffectParams struct {
	Player uuid.UUID  `json:"player,omitempty"`
	Object ObjectRef  `json:"object,omitempty"`
	Amount int        `json:"amount,omitempty"`
	Cost   string     `json:"cost,omitempty"`
	Name   string     `json:"name,omitempty"`
	Filter CastFilter `json:"filter,omitempty"`
}

// CastFilter is a spell predicate as data: the card types a spell may
// have, any one of which matches. Empty matches every spell. It is
// what WhenYouNextCast's "an instant or sorcery spell" became.
//
// Types is a CLOSED vocabulary — CR 205.2a's card types, spelled as the
// rules spell them ("Instant", "Sorcery") — and matching is EXACT
// against the card's type list, never a substring of its type line
// (#1568 review: "art" would otherwise match Artifact, and "" every
// spell). A filter naming anything else is refused where it is
// scheduled and where it is restored (ErrUnknownEffectKey).
type CastFilter struct {
	Types []string `json:"types,omitempty"`
}

// cardTypes205 is CR 205.2a's list of card types.
var cardTypes205 = map[string]bool{
	"Artifact": true, "Battle": true, "Conspiracy": true, "Creature": true,
	"Dungeon": true, "Enchantment": true, "Instant": true, "Kindred": true,
	"Land": true, "Phenomenon": true, "Plane": true, "Planeswalker": true,
	"Scheme": true, "Sorcery": true, "Vanguard": true,
}

// Valid reports whether every type is one of CR 205.2a's card types,
// spelled exactly.
func (f CastFilter) Valid() bool {
	for _, t := range f.Types {
		if !cardTypes205[t] {
			return false
		}
	}
	return true
}

// Matches reports whether the card has one of the filter's types,
// compared as whole type names. An invalid filter matches nothing.
func (f CastFilter) Matches(c Card) bool {
	if len(f.Types) == 0 {
		return true
	}
	if !f.Valid() {
		return false
	}
	have := exactCardTypesOf(c)
	for _, t := range f.Types {
		for _, h := range have {
			if strings.EqualFold(h, t) {
				return true
			}
		}
	}
	return false
}

// exactCardTypesOf is the card's type LIST — the effective one when the
// layer cache is warm, else the printed one parsed out of the type line
// — for a whole-word comparison.
func exactCardTypesOf(c Card) []string {
	if c.effective != nil {
		return c.effective.Types
	}
	if c.FaceDownIsPermanent() {
		return faceDownCharacteristic(c).Types
	}
	_, types, _ := ParseTypeLine(c.TypeLine)
	return types
}

// BodyFunc is what a delayed trigger does when its item resolves. The
// contract StackItem.Effect has always had: read everything off the
// item and the game handed in, capture nothing — and now it cannot
// capture anything, because it is a registered package-level function.
type BodyFunc func(g *Game, item *StackItem, p EffectParams) error

// ConditionFunc is an event-conditioned delayed trigger's match: a
// watched event, the trigger (for its controller) and its CondParams.
// Runs under g.mu held in write mode; must not call locking mutators.
type ConditionFunc func(ev Event, dt *DelayedTrigger, g *Game, p EffectParams) bool

// BodyRef names a registered body. The key is unexported, so the only
// way to hold one is DelayedBody / SimpleDelayedBody — which is what
// makes a func literal in a delayed trigger a compile error.
type BodyRef struct{ key string }

// Key is the body's on-disk key.
func (r BodyRef) Key() string { return r.key }

// effectKeyFault reports a programming error in how a delayed trigger
// was built — a forgotten Body, an invalid filter. In a test binary it
// panics, so the first test that schedules it fails; in production it
// logs and the caller drops the trigger, because a malformed card must
// never take the server down (#1568 review).
func effectKeyFault(msg string) {
	if testing.Testing() {
		panic(msg)
	}
	slog.Error(msg)
}

// ConditionRef names a registered event condition.
type ConditionRef struct{ key string }

// Key is the condition's on-disk key.
func (r ConditionRef) Key() string { return r.key }

// effectRegistryMu guards the three maps. Production registers only at
// init, before any game runs; tests register their own bodies while
// other tests' games resolve, so the maps need the lock to be race-free.
var effectRegistryMu sync.RWMutex

var (
	effectBodies     = map[string]BodyFunc{}
	effectConditions = map[string]ConditionFunc{}
	effectAliases    = map[string]string{}
	effectKeyPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*/[a-z0-9]+(-[a-z0-9]+)*$`)
)

func checkEffectKey(kind, key string) {
	if !effectKeyPattern.MatchString(key) {
		panic(fmt.Sprintf("game: %s key %q is not <area>/<name>, lowercase and hyphenated", kind, key))
	}
	if _, dup := effectBodies[key]; dup {
		panic(fmt.Sprintf("game: %s key %q is already registered as a body", kind, key))
	}
	if _, dup := effectConditions[key]; dup {
		panic(fmt.Sprintf("game: %s key %q is already registered as a condition", kind, key))
	}
	if _, dup := effectAliases[key]; dup {
		panic(fmt.Sprintf("game: %s key %q is already an alias", kind, key))
	}
}

// DelayedBody registers a body. Call it once, from a package-level var
// or init, in the file that defines the function. Panics on a bad or
// duplicate key, the way the catalog's Register does.
func DelayedBody(key string, fn BodyFunc) BodyRef {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("body", key)
	if fn == nil {
		panic(fmt.Sprintf("game: body %q has no function", key))
	}
	effectBodies[key] = fn
	return BodyRef{key: key}
}

// SimpleDelayedBody registers a body that reads no params — every
// "return the listed cards" and "draw a card" shape.
func SimpleDelayedBody(key string, fn func(g *Game, item *StackItem) error) BodyRef {
	if fn == nil {
		panic(fmt.Sprintf("game: body %q has no function", key))
	}
	return DelayedBody(key, func(g *Game, item *StackItem, _ EffectParams) error { return fn(g, item) })
}

// DelayedCondition registers an event condition.
func DelayedCondition(key string, fn ConditionFunc) ConditionRef {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("condition", key)
	if fn == nil {
		panic(fmt.Sprintf("game: condition %q has no function", key))
	}
	effectConditions[key] = fn
	return ConditionRef{key: key}
}

// EffectAlias keeps an old key resolving after a rename: a restore
// point written before the rename names the old key, and the ledger
// never lets it disappear.
func EffectAlias(oldKey, newKey string) {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("alias", oldKey)
	if _, ok := effectBodies[newKey]; !ok {
		if _, ok := effectConditions[newKey]; !ok {
			panic(fmt.Sprintf("game: alias %q names %q, which is not registered", oldKey, newKey))
		}
	}
	effectAliases[oldKey] = newKey
}

func resolveEffectAlias(key string) string {
	if to, ok := effectAliases[key]; ok {
		return to
	}
	return key
}

func lookupBody(key string) (BodyFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := effectBodies[resolveEffectAlias(key)]
	return fn, ok
}

func lookupCondition(key string) (ConditionFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := effectConditions[resolveEffectAlias(key)]
	return fn, ok
}

// KnownEffectBody reports whether this binary has a body for key (an
// alias counts).
func KnownEffectBody(key string) bool { _, ok := lookupBody(key); return ok }

// KnownEffectCondition reports whether this binary has a condition for
// key (an alias counts).
func KnownEffectCondition(key string) bool { _, ok := lookupCondition(key); return ok }

// RegisteredEffectKeys lists every registered key as "body <key>",
// "condition <key>" or "alias <old> <new>", sorted — for the ledger.
func RegisteredEffectKeys() []string {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	out := make([]string, 0, len(effectBodies)+len(effectConditions)+len(effectAliases))
	for k := range effectBodies {
		out = append(out, "body "+k)
	}
	for k := range effectConditions {
		out = append(out, "condition "+k)
	}
	for k, v := range effectAliases {
		out = append(out, "alias "+k+" "+v)
	}
	sort.Strings(out)
	return out
}

// bodyEffect is the stack item Effect for a keyed body. The function is
// looked up at RESOLUTION, not captured at fire time, so a restored
// item resolves through the running binary — which is the point.
//
// A key with no body is unreachable in play (registration and restore
// both refuse one); if it happens anyway the resolution is an error the
// table sees rather than a silent no-op.
func bodyEffect(key string, p EffectParams) func(g *Game, item *StackItem) error {
	return func(g *Game, item *StackItem) error {
		fn, ok := lookupBody(key)
		if !ok {
			return fmt.Errorf("%w: delayed-trigger body %q", ErrUnknownEffectKey, key)
		}
		return fn(g, item, p)
	}
}

// isZero reports whether p carries nothing — what the snapshot omits.
func (p EffectParams) isZero() bool {
	return p.Player == uuid.Nil && p.Object == (ObjectRef{}) && p.Amount == 0 &&
		p.Cost == "" && p.Name == "" && len(p.Filter.Types) == 0
}

// effectParamsOrNil is the snapshot mirror's form: a copy, or nil when
// there is nothing to write.
func effectParamsOrNil(p EffectParams) *EffectParams {
	if p.isZero() {
		return nil
	}
	c := cloneEffectParams(p)
	return &c
}

// effectParamsValue is the restore half.
func effectParamsValue(p *EffectParams) EffectParams {
	if p == nil {
		return EffectParams{}
	}
	return cloneEffectParams(*p)
}
