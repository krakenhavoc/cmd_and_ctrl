package game

import "github.com/google/uuid"

// card_ptr_reads.go — pointer twins of the by-value Card readers the
// view calls once per card (#2113).
//
// Card is over a kilobyte. A value-receiver method called through a
// *Card copies the whole card to make its receiver, inlined or not
// (#1498 measured it), and the view called about ten of them per card,
// on every frame. Each twin here does exactly what its by-value
// namesake does; the by-value form stays, as a thin wrapper where the
// body is shared and as the original where it is not, so no catalog
// file or open branch has to change.
//
// Naming: a reader that computes a value is `<Name>Of`, a predicate is
// `<Name>Ptr`.

// CatalogKeyOf is CatalogKey without the receiver copy.
func CatalogKeyOf(c *Card) string { return catalogKeyOf(c) }

// EffectiveOf is Card.Effective without the receiver copy. The slices
// in the returned characteristic are shared with the layer cache and
// are read-only.
func EffectiveOf(c *Card) Characteristic { return effectiveOf(c) }

// IsTokenPtr is Card.IsToken without the receiver copy.
func IsTokenPtr(c *Card) bool { return typeLineHas(c.TypeLine, "token") }

// IsCreaturePtr is Card.IsCreature without the receiver copy.
func IsCreaturePtr(c *Card) bool { return hasCardType(c, "creature") }

// IsBattlePtr is Card.IsBattle without the receiver copy.
func IsBattlePtr(c *Card) bool { return hasCardType(c, "battle") }

// IsAttachedPtr is Card.IsAttached without the receiver copy.
//
// The body lives here and Card.IsAttached wraps it, so the two cannot
// drift.
func IsAttachedPtr(c *Card) bool {
	switch c.AttachedTo.Kind {
	case TargetCard, TargetPlayer:
		return c.AttachedTo.ID != uuid.Nil
	}
	return false
}

// IsMeldedPtr is Card.IsMelded without the receiver copy.
func IsMeldedPtr(c *Card) bool { return len(c.MeldedFrom) > 0 }

// IsMeldCardPtr is Card.IsMeldCard without the receiver copy.
func IsMeldCardPtr(c *Card) bool { return c.Meld != nil && c.Meld.ResultOracleID != "" }
