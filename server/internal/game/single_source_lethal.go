package game

import "github.com/google/uuid"

// single_source_lethal.go — #6209: "This creature can't be destroyed by
// lethal damage unless lethal damage dealt by a single source is marked
// on it" (Ogre Enforcer).
//
// CR 704.5g is a sum: "it has damage marked on it, and the total damage
// marked on it is greater than or equal to its toughness". Ogre
// Enforcer's clause asks a question the sum cannot answer — not how much
// damage is marked but WHO marked it — so the engine has to keep the
// marked damage per source as well as its total.
//
// # What is recorded, and why not the turn's tally
//
// TurnTally.DamagedObjects already holds every damage event dealt this
// turn with its amount (#1946), and it is the wrong record for this
// clause: it is damage DEALT, and CR 704.5g reads damage MARKED. A
// creature healed to nothing has the same tally and no marked damage at
// all, so the sum of the tally would destroy an Ogre Enforcer that the
// rules leave standing.
//
// So Card.MarkedBySource carries the per-source split of
// Card.DamageMarked itself: one entry per source object that has damage
// marked on this creature right now, in the order the damage was
// marked. It is maintained wherever DamageMarked is — the damage tail,
// the sandbox's manual mark, every heal and the battlefield-exit clear —
// and never exceeds DamageMarked in total.
//
// # Removing damage
//
// Nothing in the rules attributes removed damage to a source: "remove
// all damage" removes marked damage, full stop, and a partial heal
// (CR 701.69a) removes an amount. So the split is reduced from the
// MOST RECENTLY marked backwards, which keeps the invariant (the split
// never totals more than DamageMarked) and keeps the one question the
// clause asks honest in the only direction that matters: a creature
// whose damage was healed below lethal is no longer destroyed at all, so
// no reader is left asking which source marked what.

// SourceMark is how much damage one SOURCE has marked on a creature
// right now (#6209). The source is its instance id: an object that left
// and came back is a new object (CR 400.7), and a creature carrying its
// mark is not holding a reference to it, only a name for the clause's
// "a single source".
type SourceMark struct {
	Source uuid.UUID `json:"source"`
	Amount int       `json:"amount"`
}

// noteMarkedDamageSource records that `source` has just marked `n`
// damage on c (#6209). A source of uuid.Nil is an unattributed mark (a
// hand-built event in a test, the sandbox with no source named) and is
// recorded as one too: an unattributed mark is still one source's as far
// as the clause is concerned, and dropping it would let an Ogre Enforcer
// survive damage the board can see on it.
func (c *Card) noteMarkedDamageSource(source uuid.UUID, n int) {
	if c == nil || n <= 0 {
		return
	}
	for i := range c.MarkedBySource {
		if c.MarkedBySource[i].Source == source {
			c.MarkedBySource[i].Amount += n
			return
		}
	}
	c.MarkedBySource = append(c.MarkedBySource, SourceMark{Source: source, Amount: n})
}

// releaseMarkedDamage takes `n` damage off the per-source split, most
// recently marked first (see the file comment), and drops an entry that
// reaches zero. A non-positive `n` does nothing; more than is recorded
// clears the split.
func (c *Card) releaseMarkedDamage(n int) {
	if c == nil || n <= 0 {
		return
	}
	for i := len(c.MarkedBySource) - 1; i >= 0 && n > 0; i-- {
		take := min(c.MarkedBySource[i].Amount, n)
		c.MarkedBySource[i].Amount -= take
		n -= take
		if c.MarkedBySource[i].Amount <= 0 {
			c.MarkedBySource = append(c.MarkedBySource[:i], c.MarkedBySource[i+1:]...)
		}
	}
	if len(c.MarkedBySource) == 0 {
		c.MarkedBySource = nil
	}
}

// clearMarkedDamageSources drops the whole split, for the two places
// that set DamageMarked to zero outright: the battlefield exit and
// "remove all damage".
func clearMarkedDamageSources(c *Card) {
	if c != nil {
		c.MarkedBySource = nil
	}
}

// LethalFromASingleSource reports whether some ONE source has lethal
// damage marked on `c` (CR 704.5g's threshold, deathtouch included
// through LethalDamageThreshold). The question Ogre Enforcer asks.
//
// False for a creature with no damage marked, and false when two sources
// have marked lethal damage between them but neither alone.
func LethalFromASingleSource(c *Card) bool {
	if c == nil || c.DamageMarked <= 0 {
		return false
	}
	need := c.LethalDamageThreshold()
	if need <= 0 {
		return false
	}
	for _, m := range c.MarkedBySource {
		if m.Amount >= need {
			return true
		}
	}
	return false
}

// SingleSourceLethalOnly is one "can't be destroyed by lethal damage
// unless lethal damage dealt by a single source is marked on it" static
// (#6209, Ogre Enforcer), as the catalog declares it.
//
// A struct rather than a bool for the reason DealtDamageCantBeRegenerated
// is one: the clause's own wording rides it, for the log line and the
// card's ability row.
type SingleSourceLethalOnly struct {
	// Label is the clause as printed.
	Label string
}

// CatalogSingleSourceLethalOnly returns the statics a permanent with the
// given catalog key has. carddef.go sets it from
// CardDef.SingleSourceLethalOnly; a game-package test may stub it.
var CatalogSingleSourceLethalOnly func(key string) []SingleSourceLethalOnly

// singleSourceLethalOnlyLocked reports whether `c` prints (or has been
// given) the clause. Keyed by CatalogAbilityKey, so a creature that has
// lost all its abilities is destroyed by lethal damage like any other
// (CR 613.1f).
//
// Caller must hold g.mu with fresh layers. Reads only.
func singleSourceLethalOnlyLocked(c *Card) bool {
	if c == nil || CatalogSingleSourceLethalOnly == nil {
		return false
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return false
	}
	return len(CatalogSingleSourceLethalOnly(key)) > 0
}

// lethalDamageDestroysLocked is the CR 704.5g half of the SBA's creature
// arm: whether the damage marked on `c` destroys it.
//
// It is the plain sum (CR 704.5g) unless the creature prints Ogre
// Enforcer's clause, in which case one source must have marked lethal
// damage on its own. The deathtouch SBA (CR 704.5h) is NOT affected: it
// destroys for any nonzero damage from a deathtouch source and is a rule
// of its own, which this clause says nothing about.
//
// Caller must hold g.mu with fresh layers. Reads only.
func lethalDamageDestroysLocked(c *Card) bool {
	if c == nil || c.DamageMarked < c.LethalDamageThreshold() {
		return false
	}
	if !singleSourceLethalOnlyLocked(c) {
		return true
	}
	return LethalFromASingleSource(c)
}
