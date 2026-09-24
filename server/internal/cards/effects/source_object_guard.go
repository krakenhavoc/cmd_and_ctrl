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
// Keying on the ID alone cannot tell "this" from "each creature you
// control" when the set happens to reach the source: the source's new
// object is a legitimate member of that set (CR 400.7, CR 611.2c). So
// a caller that acts on the members of a set it read off the board —
// or on a permanent a player chose from such a set at resolution —
// hands the primitive ctx.asGroupMember(), and the question is not
// asked of those targets (#1463; ADR 0018 amendment 2026-09-24,
// Decision 20). The default stays guarded on purpose: a loop that
// forgets asGroupMember skips one object — weaker than printed — while
// a "this" call site that forgot a marker would act on a stranger,
// which is the #259 direction. Chosen targets keep the default: CR
// 608.2b already makes a target that left and came back illegal.
// Mass primitives that select with a Match predicate
// (BoostUntilEOT.Match, DestroyAllMatching, …) never asked.
//
// Sites that name the source BY CONSTRUCTION — ctx.Source(),
// item.SourceCardID, a zero Target meaning "this", a "for as long as
// ~" duration, an "until ~ leaves" — ask isNewSourceObjectAsThis,
// which ignores the group mark: those are "this" whatever Context
// they were handed.

// asGroupMember returns a copy of c for acting on the members of a set
// the text names by group — "each creature you control", "untap all
// lands you control" — rather than on "this". Primitives applied with
// it do not ask the #1432 question of their target, so the source's
// new object, a member of the set like any other, is acted on. The
// copy shares Game and Item; nothing else is carried.
func (c *Context) asGroupMember() *Context {
	if c == nil {
		return nil
	}
	cp := *c
	cp.groupMember = true
	return &cp
}

// isNewSourceObject reports whether `target` is this context's own
// source and that source is on the battlefield as a new object the
// ability knows nothing about. See the file comment. False for a nil
// context, a nil item (an AsEnters hook), uuid.Nil, any target that is
// not the source, and any target reached through asGroupMember.
func (c *Context) isNewSourceObject(target uuid.UUID) bool {
	if c == nil || c.groupMember {
		return false
	}
	return c.isNewSourceObjectAsThis(target)
}

// isNewSourceObjectAsThis is isNewSourceObject for a site that names
// the source by construction ("this" whatever Context it was handed),
// so the group mark does not switch it off.
func (c *Context) isNewSourceObjectAsThis(target uuid.UUID) bool {
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
