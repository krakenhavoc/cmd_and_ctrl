package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_object_guard.go — "this permanent" is an OBJECT, not a card
// (#1432, CR 400.7; ADR 0018 amendment 2026-09-24, Decisions 17-19).
//
// An ability that says "put a +1/+1 counter on this creature",
// "sacrifice it" or "untap this artifact" names the permanent it came
// from. Card files spell that `Target: item.SourceCardID`, and an
// instance ID survives a zone change: a creature flickered in response
// comes back under the same ID as a NEW object with no memory of the
// ability. Acting on it anyway makes the card stronger than printed —
// the #259 direction.
//
// The fix lives HERE, in the primitives, not in the ~100 card files
// that pass the source's ID. Every primitive that acts on a permanent
// asks isNewSourceObject about its target before it acts, and does
// nothing when the answer is yes. A card file cannot forget to ask,
// because it never asks: it hands the primitive an ID and the
// primitive knows whether that ID is its own item's source.
//
// What the question is (game.AbilitySourceIsNewObjectForEffect):
//
//   - the target is the resolving item's own SOURCE card, and
//   - the item is a stamped ability (StackItem.SourceObject, #1418),
//     never a spell, and
//   - that card is on the battlefield now as a different object, and
//   - this resolution did not put it there itself.
//
// The last two are the exceptions, and they are one rule: the guard
// is about a permanent that left and came back WITHOUT this effect's
// help. Text that follows the card to another zone — "when this dies,
// return it to its owner's hand", "return this card from your
// graveyard", "shuffle it into its owner's library" — is not asked
// about, because the card is not on the battlefield. Text that acts on
// the object its own effect just moved — unearth's "it gains haste", a
// self-flicker's "return it with a +1/+1 counter" — is not caught,
// because the move happened during this resolution.
//
// The cost of keying on the ID is one corner, stated rather than
// hidden: a card that loops over "each creature you control" with a
// per-permanent primitive, and whose own source was flickered in
// response to that very ability, skips the new object. That is weaker
// than printed, never stronger, and it needs the source to be both a
// member of its own "each" set and a new object — mass primitives that
// select with a Match predicate (BoostUntilEOT.Match,
// DestroyAllMatching, …) are not asked, so they are not affected.

// isNewSourceObject reports whether `target` is this context's own
// source and that source is on the battlefield as a new object the
// ability knows nothing about. See the file comment. False for a nil
// context, a nil item (an AsEnters hook), uuid.Nil, and any target
// that is not the source.
func (c *Context) isNewSourceObject(target uuid.UUID) bool {
	if c == nil || c.Item == nil || c.Game == nil || target == uuid.Nil || target != c.Item.SourceCardID {
		return false
	}
	return c.Game.AbilitySourceIsNewObjectForEffect(c.Item)
}

// sourceIsNewObject is isNewSourceObject for a body written against
// (g, item) — the shape every triggered and activated Effect has — and
// the check a card file makes before it calls a game mutator on its
// own source directly rather than through a primitive.
func sourceIsNewObject(g *game.Game, item *game.StackItem) bool {
	return g != nil && g.AbilitySourceIsNewObjectForEffect(item)
}

// withoutNewSourceObject is `ids` minus this context's source when
// the source is a new object — for the primitives that take a list
// (TurnFaceDown, PhaseOut). Returns `ids` itself when nothing is
// dropped, so the common case allocates nothing.
func (c *Context) withoutNewSourceObject(ids []uuid.UUID) []uuid.UUID {
	for i, id := range ids {
		if !c.isNewSourceObject(id) {
			continue
		}
		out := append([]uuid.UUID(nil), ids[:i]...)
		for _, rest := range ids[i+1:] {
			if rest != id {
				out = append(out, rest)
			}
		}
		return out
	}
	return ids
}
