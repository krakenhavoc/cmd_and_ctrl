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

// Paid is what the announcement that put this item on the stack
// actually cost (#789 / #761): the mana tokens that left the pool,
// the counters removed or added, the life paid — and whether the
// engine charged at all (PaidCost.OnPaper).
//
// The accessors below are the readable half; this is here for a card
// that needs more than one fact about the payment.
func (c *Context) Paid() game.PaidCost {
	if c.Item == nil {
		return game.PaidCost{}
	}
	return c.Item.Paid
}

// CountersRemoved is how many counters the ability's cost removed —
// "for each counter removed this way". A fact about the ANNOUNCEMENT,
// read back the way X is: the counters are off the board by
// resolution, so nothing here could recompute it. Added in #789.
func (c *Context) CountersRemoved() int {
	return c.Paid().CountersRemoved
}

// Sacrificed is how many permanents the announcement's cost
// sacrificed — "for each artifact sacrificed this way" (Radiant
// Lotus). A fact about the ANNOUNCEMENT, read back the way X and
// CountersRemoved are: by resolution the permanents are in
// graveyards, so nothing here could recompute it. Added in #1213.
//
// For an ability whose clause prints a fixed count this is simply
// that number; the reason it exists is the VARIABLE count
// ("sacrifice one or more", "sacrifice X"), where the card itself
// cannot say.
func (c *Context) Sacrificed() int {
	return c.Paid().Sacrificed
}

// ReturnedAttacking is what the permanent the announcement's
// return-to-hand cost returned was ATTACKING — the player,
// planeswalker or battle, uuid.Nil when it was attacking nothing.
// A fact about the ANNOUNCEMENT, read back the way X, CountersRemoved
// and Sacrificed are, and for a stronger reason than any of them: the
// returned permanent's exit CLEARS Card.AttackingTarget and LKI
// carries no combat state, so by resolution there is nowhere else the
// answer could come from. Added in #1227.
//
// Ninjutsu is the clause that asks (CR 702.49a): the ninja is put
// onto the battlefield attacking the same player or planeswalker the
// returned creature was attacking.
func (c *Context) ReturnedAttacking() uuid.UUID {
	return c.Paid().ReturnedAttacking
}

// TappedPower is the power of the permanent the announcement's
// tap-another cost tapped (#759) — station's "charge counters equal
// to the tapped creature's power" (CR 702.184a). ok is false when
// the cost tapped nothing, which is every ability without the
// component.
//
// NOT a number off the payment record: CR 608.2h reads the tapped
// creature AS THE ABILITY RESOLVES if it is still on the battlefield
// (pump it in response and more counters go on), and as it last
// existed there if it is not — game.PaidTapPowerForEffect answers
// both. The one printed clause taps a single permanent; for a cost
// that taps more this is the first one named.
//
// Not clamped: a negative power is returned as negative, and the
// caller decides what "that many counters" means for it.
func (c *Context) TappedPower() (power int, ok bool) {
	taps := c.Paid().TappedOthers
	if len(taps) == 0 || c.Game == nil {
		return 0, false
	}
	return c.Game.PaidTapPowerForEffect(taps[0]), true
}

// ManaSpent is what the payment for THIS stack item can be asked
// about: Colors(), Count("R"), Total(), FromTreasure(), Snow() and
// the rest of game.ManaSpent's vocabulary.
//
// The spell's own record, so this is the read a RESOLVING spell makes
// ("if {U} was spent to cast this spell, draw a card" — Ribbons of
// Night). A permanent asking the same question about the spell it
// came from wants ManaSpentToCastThis below; the two return the same
// type on purpose.
//
// Empty for an ability item and for a copy (CR 707.10), and the
// weaker-than-printed answer to everything for a payment the engine
// waived — ask .Known() to tell the last case apart. Returned by
// value since #1212; was []game.ManaToken in #761.
func (c *Context) ManaSpent() game.ManaSpent {
	return c.Paid().Spent()
}

// ManaSpentToCastThis is CR 400.7d's half of the same question: what
// paid for the spell that became the PERMANENT this effect is running
// for.
//
// The sibling of Escaped() below, and it exists for the same reason:
// "when this creature enters, if mana from a Treasure was spent to
// cast it" resolves as a TRIGGER, after the spell has finished
// resolving, so ManaSpent above answers about the trigger's own
// (empty) record and the fact lives on the permanent instead.
//
// The zero view for an ability whose source is not on the
// battlefield, which is also the honest answer: CR 400.7d is written
// about a permanent. Added in #1212.
func (c *Context) ManaSpentToCastThis() game.ManaSpent {
	return c.CastProvenance().Spent()
}

// ColorsSpent is the distinct COLOURS of mana spent to cast this
// spell, in WUBRG order — converge's X (CR 702.86) and sunburst's
// counter count (CR 702.44). Colourless is not a colour, so {C} never
// appears. Empty when the payment was not recorded, which is the
// weaker-than-printed answer. Added in #761.
func (c *Context) ColorsSpent() []string {
	return c.Paid().ColorsSpent()
}

// ColorsSpentCount is len(ColorsSpent) — the number converge and
// sunburst actually want. Added in #761.
func (c *Context) ColorsSpentCount() int {
	return c.Paid().ColorsSpentCount()
}

// ManaSpentOfColor is how many mana of one colour paid for this
// spell: adamant's "if at least three red mana was spent to cast
// this spell" is ManaSpentOfColor("R") >= 3. Added in #761.
func (c *Context) ManaSpentOfColor(color string) int {
	return c.Paid().SpentOfColor(color)
}

// NoManaSpent reports "if no mana was spent to cast it" (Vexing
// Bauble, Satoru, the Infiltrator).
//
// True only when the engine KNOWS nothing was spent — a free cast, a
// {0} alternative cost, a copy of a spell. A payment the engine
// waived (permissive mode, ForceCast) answers false: the player paid
// something we did not see, and a punisher that fired on it would
// counter half the spells cast at a permissive table. Added in #761.
func (c *Context) NoManaSpent() bool {
	return c.Paid().NoManaSpent()
}

// ManaSpentKnown reports whether the mana half of the record is a
// fact rather than a waived charge. For a card that wants to say
// "unknown" out loud instead of folding it into the weaker answer.
// Added in #761.
func (c *Context) ManaSpentKnown() bool {
	return c.Paid().Known()
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

// Modes returns the announce-time mode indexes of a modal spell or
// ability (empty for non-modal cards), in announce order and with
// repeats when the card allows them (CR 700.2d). Added in S20 sub-PR
// 4; a multiset since #764.
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

// ModeCount is how many times option i was chosen (CR 700.2d). 0
// when the option was not chosen, 1 for the ordinary modal card, and
// more only for a Repeatable ModeSpec — Mystic Confluence's draw
// mode taken three times is 3. Added by #764.
func (c *Context) ModeCount(i int) int {
	return game.ModeCount(c.Modes(), i)
}

// Mode is the OPTION index chosen at occurrence `n` of the announced
// mode list, or -1 when there is no such occurrence. Modes resolve in
// announce order (CR 700.2c), so a card that walks its occurrences
// walks this. Added by #764.
func (c *Context) Mode(n int) int {
	modes := c.Modes()
	if n < 0 || n >= len(modes) {
		return -1
	}
	return modes[n]
}

// ModeTargets is the target group announced for occurrence `n` — the
// picks that answered THAT occurrence's clauses, in clause order.
// A repeated mode's two occurrences have two groups, which is what
// makes CR 700.2d work. Added by #764.
func (c *Context) ModeTargets(n int) []game.TargetRef {
	var out []game.TargetRef
	for _, t := range c.Targets() {
		if t.Mode == n {
			out = append(out, t)
		}
	}
	return out
}

// ClauseTargets is the picks that answered clause `slot` of the
// announcement — slot 0 is "target creature you control", slot 1 is
// "target creature or planeswalker you don't control". For a modal
// item, narrow to one occurrence with ModeTargets first. Added by
// #764.
func (c *Context) ClauseTargets(slot int) []game.TargetRef {
	var out []game.TargetRef
	for _, t := range c.Targets() {
		if t.Slot == slot {
			out = append(out, t)
		}
	}
	return out
}

// ClauseTarget is the single pick that answered clause `slot`, and
// whether there is one and it is still legal (CR 608.2b). The read a
// two-slot positional card wants: "the creature I control" is
// ClauseTarget(0), "the thing it hits" is ClauseTarget(1). Added by
// #764.
func (c *Context) ClauseTarget(slot int) (game.TargetRef, bool) {
	for _, t := range c.Targets() {
		if t.Slot != slot {
			continue
		}
		return t, c.IsTargetLegal(t)
	}
	return game.TargetRef{}, false
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

// OptionalCostTimes is how many times the optional additional cost
// keyed `key` was paid when this spell was announced (CR 601.2b,
// ADR 0073). Zero when it was declined, when the card offers no such
// cost, and for every ability item.
//
// Keyed rather than indexed so a card's resolution never has to know
// its own declaration order, exactly as PaidAltCost is keyed.
func (c *Context) OptionalCostTimes(key string) int {
	if c.Item == nil || len(c.Item.Paid.OptionalCosts) == 0 {
		return 0
	}
	card, ok := c.Game.LookupCardForEffect(c.Item.ID)
	if !ok {
		return 0
	}
	return game.OptionalCostTimesPaid(card, c.Item.Paid.OptionalCosts, key)
}

// WasKicked is CR 702.33's "if this spell was kicked" — the read a
// kicked spell's own resolution branches on, the same shape
// PaidAltCost gives an overloaded one:
//
//	if ctx.WasKicked() { … 4 damage … } else { … 2 damage … }
//
// True for kicker and multikicker alike: no rules text tells them
// apart, and no card prints both.
func (c *Context) WasKicked() bool { return c.KickedTimes() > 0 }

// KickedTimes is CR 702.33d's "the number of times it was kicked" —
// Wolfbriar Elemental's Wolf count. 1 for an ordinary kicked spell,
// 0 for an unkicked one.
func (c *Context) KickedTimes() int {
	return c.OptionalCostTimes(game.KickerKey) + c.OptionalCostTimes(game.MultikickerKey)
}

// CastProvenance is what the permanent this effect is running for
// remembers about the SPELL it came from — CR 400.7d, "an ability of
// a permanent can reference information about the spell that became
// that permanent as it resolved, including what costs were paid".
//
// The SOURCE's record, not the item's. PaidAltCost above answers "what
// paid for the spell I am resolving", which is the right question for
// an overloaded Cyclonic Rift and the wrong one for Phlage: by the
// time "sacrifice it unless it escaped" resolves, the item on the
// stack is the TRIGGER, and the spell that paid escape finished
// resolving two steps ago. The fact lives on the permanent from #653.
//
// The zero record for an ability whose source is not on the
// battlefield, which is also the honest answer: CR 400.7d is written
// about a permanent.
func (c *Context) CastProvenance() game.CastProvenance {
	if c.Game == nil || c.Item == nil {
		return game.CastProvenance{}
	}
	return c.Game.CastProvenanceForEffect(c.Item.SourceCardID)
}

// Escaped is CR 702.138b for the permanent this effect is running for:
// it was cast for its escape cost and is still the permanent that
// spell became. The whole of "sacrifice it unless it escaped".
//
// False for a reanimated, hard-cast, blinked or put-onto-the-
// battlefield copy of the same card, and false is the
// weaker-than-printed answer in every one of those cases.
func (c *Context) Escaped() bool {
	return c.CastProvenance().Escaped()
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

// Trigger is the event that fired this triggered ability (#1223): the
// amount for a damage or life event, the CR 603.10 last-known
// snapshot of the object a zone change was about, the counters
// placed, the spell cast, the ability activated.
//
//	ctx.Trigger().Event.Amount      // "that much damage"
//	ctx.Trigger().Object.ManaValue  // "the creature that died"
//
// The zero value — `Fired()` false, `Object` nil — for a spell, an
// activated ability, a CR 603.12 reflexive trigger and a trigger
// restored from a snapshot written before the field existed. A card
// that reads it unconditionally gets zeroes rather than a panic,
// which is the #259 direction: the weaker-than-printed answer.
//
// Read this rather than walking `ctx.Game.Events` backwards for the
// event. The log scan finds the LAST event of a kind, which is the
// wrong one the moment two land in the same batch (two creatures
// dying to one wrath), and it finds nothing at all for a trigger that
// resolved a priority round later.
func (c *Context) Trigger() game.TriggerContext {
	if c.Item == nil || c.Item.Trigger == nil {
		return game.TriggerContext{}
	}
	return *c.Item.Trigger
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

// ChosenPlayer is the player most recently named by a
// ChoosePlayer on this item (#929, choose_player.go), or uuid.Nil
// when the last question could not be asked and when none was asked
// at all.
//
// A chosen player is NOT a target: it is picked while the effect
// resolves, nothing may respond to it, and nothing re-checks it
// against the board. A branch that acts on a seat which may since
// have left checks for itself, exactly as Payload's readers do.
func (c *Context) ChosenPlayer() uuid.UUID {
	return game.ChosenPlayerOn(c.Item)
}

// ChosenPlayers is every player named by a ChoosePlayer on this item,
// in the order the card asked. The read behind "choose a SECOND
// player": pass it back as ChoosePlayer.Except.
func (c *Context) ChosenPlayers() []uuid.UUID {
	return game.ChosenPlayersOn(c.Item)
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
