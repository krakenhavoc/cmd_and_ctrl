package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// life_cost_count.go — #1594, ADR 0020 Decision 47: a life payment
// whose AMOUNT is computed when the ability is activated.
//
//	War Room           {3}, {T}, Pay life equal to the number of colors
//	                   in your commanders' color identity: Draw a card.
//	Murderous Betrayal {B}{B}, Pay half your life, rounded up: Destroy
//	                   target nonblack creature. …
//
// AbilityCost.Life is a printed number. These costs print a rule for
// the number instead, and CR 601.2f–g (through CR 602.2b) says when the
// rule is read: the total cost is determined, and LOCKED IN, before any
// of it is paid. So the count is read once, at announce, by
// ActivateCatalogAbility, and the amount it returns is the amount
// charged and recorded on PaidCost.LifePaid. Nothing re-reads it: a
// commander that leaves, or a life total that moves, after the
// activation changes nothing about what was paid.
//
// The count is a KEY, not a func, for the reason ModeCountCondition is
// (modes.go): an AbilityCost is reachable from Game through every
// card's ActivatedAbilities, and ADR 0041 phase 3's closure ratchet
// (testdata/closure_fields.txt) admits no new func-typed route. The
// function lives in a registry here, registered once at init by
// LifeCount, and the cost carries only its name — so the cost survives
// Clone, undo and a snapshot restore as data, and the restored game
// finds the same function by the same key.
//
// Three readers, one function (AbilityLifeCostLocked), so they cannot
// disagree:
//
//   - the engine's announce gate and payment (activated.go);
//   - the legal-move enumerator, which drops an activation whose
//     computed amount the seat cannot pay (CR 119.4, the #695 rule:
//     never offer a life payment the engine would refuse) and puts the
//     computed amount on Move.Cost.Life, which is what a bot prices;
//   - the view, which stamps the computed amount onto the ability row's
//     life_cost — the "pay N life" chip the client renders.
//
// Out of scope, deliberately: "Pay X life" on an ability (Krumar
// Initiate). That amount is the announced X, and the life total is a
// CEILING on X — the shape Toxic Deluge's AdditionalCost.PayLifeX has
// on the cast path, where the enumerator solves X against life. A count
// that took X as an argument would compute the payment but leave the
// X search blind to it.

// LifeCostCount names a registered "pay life equal to <count>" rule.
// The key is unexported, so the only way to hold a non-zero one is
// LifeCount — a func literal on an AbilityCost does not compile.
type LifeCostCount struct{ key string }

// Key is the count's registry key.
func (c LifeCostCount) Key() string { return c.key }

// IsZero reports whether no count is named.
func (c LifeCostCount) IsZero() bool { return c.key == "" }

// LifeCountFunc computes a life payment's amount. `activator` is the
// player activating the ability — "you", "your life", "your
// commanders" — and `source` is the object whose ability it is. It is
// read-only and runs under g.mu, so it must not call a public locking
// accessor. A negative answer is read as 0.
type LifeCountFunc func(g *Game, activator, source uuid.UUID) int

var lifeCostCounts = struct {
	sync.RWMutex
	byKey map[string]LifeCountFunc
}{byKey: map[string]LifeCountFunc{}}

// LifeCount registers a life-payment count under `key` and returns its
// name. Call it once, from a package-level var. Panics on an empty key,
// a nil function or a duplicate, each of which is a card-file bug that
// would otherwise ship a cost the engine cannot price.
func LifeCount(key string, fn LifeCountFunc) LifeCostCount {
	if key == "" {
		panic("game: LifeCount with an empty key")
	}
	if fn == nil {
		panic(fmt.Sprintf("game: life count %q has no function", key))
	}
	lifeCostCounts.Lock()
	defer lifeCostCounts.Unlock()
	if _, dup := lifeCostCounts.byKey[key]; dup {
		panic(fmt.Sprintf("game: life count %q registered twice", key))
	}
	lifeCostCounts.byKey[key] = fn
	return LifeCostCount{key: key}
}

// AbilityLifeCostLocked is the life an activation of `cost` by
// `activator` from `source` charges right now: the printed Life plus
// the named count, if any. ok is false only for a count whose key is
// not registered — impossible for a cost built through LifeCount, and
// read by every caller as "this cost cannot be paid" rather than as
// "it is free", the weaker reading of a cost the engine cannot price.
//
// Caller must hold g.mu (read or write).
func (g *Game) AbilityLifeCostLocked(activator, source uuid.UUID, cost AbilityCost) (life int, ok bool) {
	life = cost.Life
	if cost.LifeFrom.IsZero() {
		return life, true
	}
	lifeCostCounts.RLock()
	fn, found := lifeCostCounts.byKey[cost.LifeFrom.key]
	lifeCostCounts.RUnlock()
	if !found {
		return 0, false
	}
	if n := fn(g, activator, source); n > 0 {
		life += n
	}
	return life, true
}

// CommanderIdentityColorCountForEffect is the number of colours in
// `player`'s commanders' combined colour identity (CR 903.4) — War
// Room's count. Every commander the player OWNS contributes, wherever
// it is (commanderIdentityFor's search), so a partner pair counts the
// union of both halves and a background pairing its commander's too.
// No commander, and a colourless one, are both 0.
//
// Caller must hold g.mu.
func (g *Game) CommanderIdentityColorCountForEffect(player uuid.UUID) int {
	return len(commanderIdentityFor(g, g.playerByIDLocked(player)).Colors)
}
