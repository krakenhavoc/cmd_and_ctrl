package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Context is the scoped API surface handed to a Spec's OnResolve /
// AsEnters callback. Wraps the Game (under write-lock held by the
// resolution path) and the stack item currently being resolved. An
// OnResolve that needs to read game state or mutate it goes through
// Context rather than through *game.Game directly, so the hard rule
// "primitives run under the already-held write lock; never call a
// public locking mutator" stays visible per method.
//
// NB: Go can't encode "locked context" at the type level. The
// guardrail is documentation + code review. A primitive that reaches
// into Context.Game and calls g.CastSpell (or any other public
// method that takes g.mu itself) will deadlock. Stay on the
// ForEffect methods declared in server/internal/game/effect_api.go.
type Context struct {
	// Game is the live game under resolution. Caller holds g.mu
	// write — read fields directly, mutate via *ForEffect methods.
	Game *game.Game

	// Item is the stack item being resolved (spell or ability).
	// Nil for AsEnters callbacks (which fire after the spell has
	// already resolved and routed to the battlefield).
	Item *game.StackItem
}

// NewContext constructs a Context bound to a game + stack item.
// Called from the resolution path just before OnResolve / AsEnters
// fires.
func NewContext(g *game.Game, item *game.StackItem) *Context {
	return &Context{Game: g, Item: item}
}

// Controller returns the caster / controller of the current stack
// item. uuid.Nil for AsEnters contexts with no item (the resolution
// path can pass a synthetic item with Controller set to the owner
// of the entering permanent if a primitive needs this).
func (c *Context) Controller() uuid.UUID {
	if c.Item == nil {
		return uuid.Nil
	}
	return c.Item.Controller
}

// Source returns the ID of the card producing the current effect
// (the spell itself for spell resolution, the source permanent for
// activated / triggered abilities). Populated into emitted events
// so downstream consumers can attribute "who did what."
func (c *Context) Source() uuid.UUID {
	if c.Item == nil {
		return uuid.Nil
	}
	return c.Item.SourceCardID
}

// X returns the announce-time value of X for the current stack
// item (0 when the spell has no X or none was announced). The cost
// engine already charged X·generic at cast time; effects read it
// here to scale damage / draw / life. Added in S20 sub-PR 3.
func (c *Context) X() int {
	if c.Item == nil || c.Item.XValue < 0 {
		return 0
	}
	return c.Item.XValue
}

// Opponents returns the IDs of every seated, non-eliminated player
// other than the current item's controller, in seat order. "Each
// opponent" effects (Exsanguinate) iterate this. Added in S20
// sub-PR 3.
func (c *Context) Opponents() []uuid.UUID {
	me := c.Controller()
	var out []uuid.UUID
	for _, p := range c.Game.Seats {
		if p == nil || p.Eliminated || p.ID == me {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

// Modes returns the announce-time mode indexes of a modal spell
// (empty for non-modal cards). Added in S20 sub-PR 4.
func (c *Context) Modes() []int {
	if c.Item == nil {
		return nil
	}
	return c.Item.Modes
}

// HasMode reports whether option i of the spell's ModeSpec was
// chosen at announce. A modal card's OnResolve is a sequence of
// `if ctx.HasMode(0) { … }` blocks in option order (CR 700.2c:
// modes resolve in printed order). Added in S20 sub-PR 4.
func (c *Context) HasMode(i int) bool {
	for _, m := range c.Modes() {
		if m == i {
			return true
		}
	}
	return false
}

// PaidAltCost reports whether the spell was cast for the named
// alternative cost (CR 118.9) — "overload", "evoke", "cleave". False
// for an ordinary cast and for every ability item. Added in S22.
//
// A card whose text changes with the cost is a run of
// `if ctx.PaidAltCost("overload") { … sweep … }` against the printed
// single-target branch — the same shape HasMode gives a modal card,
// and for the same reason: the choice was made at announce and the
// resolution just reads it back.
func (c *Context) PaidAltCost(key string) bool {
	return c.Item != nil && c.Item.AltCost == key && key != ""
}

// Targets returns the announce-time target slots. Callers that
// assume a specific cardinality should bounds-check — effects run
// in sandbox-adjacent territory where the UI might send too few
// or too many slots, and the primitive silently no-ops rather
// than panicking.
func (c *Context) Targets() []game.TargetRef {
	if c.Item == nil {
		return nil
	}
	return c.Item.Targets
}

// Payload is what the effect that CREATED this item told it — only a
// CR 603.12 reflexive trigger has one (see reflexive.go), and it is
// empty for everything else. Unlike Targets it is never re-checked
// against the board: it is a record of what happened, so an effect
// that cares whether the card is still there asks.
func (c *Context) Payload() []game.TargetRef {
	if c.Item == nil {
		return nil
	}
	return c.Item.Payload
}

// PayloadCards is Payload narrowed to its card refs, in the order the
// creating effect listed them. The common read: "the creatures that
// were tapped this way", "the cards revealed this way".
func (c *Context) PayloadCards() []uuid.UUID {
	var out []uuid.UUID
	for _, ref := range c.Payload() {
		if ref.Kind == game.TargetCard {
			out = append(out, ref.ID)
		}
	}
	return out
}

// PlayerByID is a read-through to the live game. Returns nil if
// the player is not seated / has left.
func (c *Context) PlayerByID(id uuid.UUID) *game.Player {
	return c.Game.PlayerByIDForEffect(id)
}

// LegalTargets returns the announce-time target slots that are still
// legal now, in announce order (S20 sub-PR 5). A multi-target effect
// iterates this instead of Targets() so a slot whose target left or
// stopped qualifying is skipped — CR 608.2b's "does as much as it
// can" — while the all-illegal fizzle has already run before
// OnResolve. Positional effects (Arc Trail) index Targets() and check
// IsTargetLegal per slot instead, since skipping would shift the
// slots.
func (c *Context) LegalTargets() []game.TargetRef {
	var out []game.TargetRef
	for _, t := range c.Targets() {
		if c.IsTargetLegal(t) {
			out = append(out, t)
		}
	}
	return out
}

// IsTargetLegal reruns the CR 608.2b check for a single target ref
// under the item's announced clause (existence only for items with
// no structured spec). Primitives use this when they want to silently skip
// a now-illegal slot (the "partial targets illegal" case) without
// aborting the whole resolution. The "all targets illegal" short-
// circuit runs before OnResolve even fires, so primitives don't
// need to handle the all-fizzle case themselves.
func (c *Context) IsTargetLegal(t game.TargetRef) bool {
	return c.Game.TargetStillLegalForEffect(c.Item, t)
}

// CreatureIDs returns the InstanceIDs of every creature currently
// on the battlefield. Iterated primitives (Wrath of God, Pyroclasm)
// call this then loop a per-card primitive over the slice. The
// snapshot is stable within the resolution — primitives do not
// mutate between rounds of iteration.
func (c *Context) CreatureIDs() []uuid.UUID {
	cards := c.Game.BattlefieldCardsForEffect()
	out := make([]uuid.UUID, 0, len(cards))
	for _, cd := range cards {
		if cd.IsCreature() {
			out = append(out, cd.InstanceID)
		}
	}
	return out
}
