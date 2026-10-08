package game

// CounterRemovalLock is a static effect that keeps one kind of counter
// from being removed from a filtered set of permanents ("Stun counters
// can't be removed from permanents your opponents control" — Fear of
// Sleep Paralysis; issue #1824, ADR 0058's 2026-10-08 amendment).
//
// It is a prohibition, so it wins over every "remove" instruction
// (CR 101.2): the counter stays and the removal does nothing. The one
// choke point is applyCounterByLocked, which every counter removal in
// the engine reaches — a stun counter replacing an untap (CR 122.1d),
// a removal effect, and a counter-removal cost's payment. A cost that
// would only be payable by removing a locked counter is refused before
// anything is paid (CR 118.3), by counterKindsPaying and
// validateCounterRemovalLocked, so a player is never charged for a
// removal that cannot happen.
//
// Derived, never stored: asked at the moment of removal, so the source
// leaving play lifts the lock with no bookkeeping.
type CounterRemovalLock struct {
	// Counter is the counter kind that cannot be removed. Empty locks
	// every kind.
	Counter string
	// Locks reports whether `target`, a permanent on the battlefield,
	// is protected from having that counter removed while `source` is
	// on the battlefield. Runs under g.mu; must not take locks.
	Locks func(target *Card, g *Game, source *Card) bool
	// Label names the clause on the card's ability rows.
	Label string
}

// CatalogCounterRemovalLocks returns the locks a catalog card declares.
// Populated by carddef.go; nil hook means none.
var CatalogCounterRemovalLocks func(oracleID string) []CounterRemovalLock

// counterRemovalLockedLocked reports whether counters of `kind` cannot
// currently be removed from c. Caller holds g.mu.
func (g *Game) counterRemovalLockedLocked(c *Card, kind string) bool {
	if c == nil || g.Battlefield == nil || CatalogCounterRemovalLocks == nil {
		return false
	}
	if !g.Battlefield.Contains(c.InstanceID) {
		return false
	}
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(source)
		if key == "" {
			continue
		}
		for _, lock := range CatalogCounterRemovalLocks(key) {
			if lock.Locks == nil || (lock.Counter != "" && lock.Counter != kind) {
				continue
			}
			if lock.Locks(c, g, source) {
				return true
			}
		}
	}
	return false
}
