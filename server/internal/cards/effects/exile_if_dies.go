package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_if_dies.go — the catalog's vocabulary for ADR 0108 §1 and §2
// (#1886, #1887): "if it would die this turn, exile it instead" and
// "it can't be regenerated this turn", and the per-recipient damage
// continuation the "dealt damage this way" forms are built on. The
// engine half is game/scoped_replacements.go (exileIfWouldDie) and
// game/regeneration.go (cantBeRegenerated).
//
// TWO FORMS, BY THE PRINTED WORDING (ADR 0108 §1 decision 3):
//
//	"If that creature would die this turn, exile it instead."   // Lava Coil
//	"If it's a creature, … if it would die this turn, exile it" // Disintegrate
//	    → ExileIfItWouldDieThisTurn{Target: id}: registered whether or
//	      not the damage was dealt. The Disintegrate ruling: "It works
//	      even if the damage is prevented or redirected."
//
//	"If a creature dealt damage this way would die this turn …"  // Demonfire
//	    → DealDamageThen(ctx, id, n, ExileIfDealtDamageWouldDie(item, false)):
//	      registered from the damage's continuation, only on a creature
//	      that took more than 0 damage.
//
// The same split holds for "can't be regenerated": "It can't be
// regenerated this turn" (Engulfing Flames) is CantBeRegeneratedThisTurn,
// and "A creature dealt damage this way can't be regenerated this turn"
// (Incinerate) is CantBeRegeneratedIfDealtDamage.
//
//	"If a creature would die this turn, exile it instead."      // Flaying Tendrils
//	    → ExileIfCreaturesWouldDieThisTurn{}: a live set (CR 611.2c).

// ExileIfItWouldDieThisTurn is "if <Target> would die this turn, exile
// it instead" (CR 614.1a, 700.4), on one permanent, until cleanup. A
// target that is not on the battlefield gets nothing.
type ExileIfItWouldDieThisTurn struct {
	Target uuid.UUID
	// Label is attribution for the log; defaults to the clause.
	Label string
}

func (e ExileIfItWouldDieThisTurn) Apply(ctx *Context) error {
	// #1432: "this creature" after a flicker in response is a new object.
	if ctx.isNewSourceObject(e.Target) {
		return nil
	}
	label := e.Label
	if label == "" {
		label = "Exiled instead if it would die this turn"
	}
	ctx.Game.ExileIfItWouldDieThisTurnForEffect(ctx.Source(), e.Target, ctx.Controller(), label)
	return nil
}

// ExileIfCreaturesWouldDieThisTurn is "If a creature would die this turn,
// exile it instead" (Flaying Tendrils, Malicious Malfunction) — or, with
// OpponentsOnly, "If a creature an opponent controls would die this
// turn, exile it instead" (Malicious Eclipse), the opponents being the
// resolving object's controller's. The set is read as each creature
// would die, so it reaches creatures that enter later this turn
// (CR 611.2c).
type ExileIfCreaturesWouldDieThisTurn struct {
	OpponentsOnly bool
	Label         string
}

func (e ExileIfCreaturesWouldDieThisTurn) Apply(ctx *Context) error {
	scope, label := game.ScopeCreatures, "If a creature would die this turn, exile it instead"
	if e.OpponentsOnly {
		scope, label = game.ScopeOpponentsCreatures, "If a creature an opponent controls would die this turn, exile it instead"
	}
	if e.Label != "" {
		label = e.Label
	}
	ctx.Game.ExileIfCreaturesWouldDieThisTurnForEffect(ctx.Source(), ctx.Controller(), scope, label)
	return nil
}

// CantBeRegeneratedThisTurn is "<Target> can't be regenerated this turn"
// (CR 701.19c): until cleanup, no regeneration shield on it is applied
// and no static regeneration replaces its destruction.
type CantBeRegeneratedThisTurn struct {
	Target uuid.UUID
	Label  string
}

func (c CantBeRegeneratedThisTurn) Apply(ctx *Context) error {
	// #1432: "this creature can't be regenerated" after a flicker in
	// response is a new object (Clergy of the Holy Nimbus).
	if ctx.isNewSourceObject(c.Target) {
		return nil
	}
	label := c.Label
	if label == "" {
		label = "Can't be regenerated this turn"
	}
	ctx.Game.CantBeRegeneratedThisTurnForEffect(ctx.Source(), c.Target, label)
	return nil
}

// RecipientThen is a damage instruction's per-recipient continuation,
// as data (#4253): a registered tail (recipientTail) told each recipient
// as TailOutcome.IDs[0] and how much it was actually dealt as its Amount,
// once that damage has settled (game.DealDamageEachEachKeyedForEffect), so
// a CR 616 prompt the damage pauses on is a restore point. The zero value
// is no continuation. Several riders on one damage chain through Next
// (AllThen).
type RecipientThen = game.KeyedTail

// recipientTail registers a per-recipient body under key, an on-disk
// identity: never renamed, never reused, ledgered "tail <key>". The body
// runs against the Context rebuilt from the tail (CardTail), told the
// recipient and what it took; the tail's Next, when set, is the next
// rider, told the same. Call it once, from a package-level var.
func recipientTail(key string, body func(ctx *Context, target uuid.UUID, dealt int, o game.TailOutcome) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		target := uuid.Nil
		if len(o.IDs) > 0 {
			target = o.IDs[0]
		}
		if err := body(ctx, target, o.Amount, o); err != nil {
			return err
		}
		if o.Tail.Next != nil {
			return o.Tail.Next.RunTold(ctx.Game, o.IDs, o.Amount)
		}
		return nil
	})
}

// itemRecipient is a per-recipient tail for ref with the resolving item
// carried as data, for a body that reads the item (the window its source
// had, CR 400.7).
func itemRecipient(item *game.StackItem, ref game.TailRef) RecipientThen {
	return NewContext(nil, item).ItemTail(ref)
}

// ExileIfDealtDamageWouldDie is "If a creature dealt damage this way
// would die this turn, exile it instead" as a continuation: a recipient
// that was dealt more than 0 damage, and is a creature on the battlefield
// — any permanent, with permanents set ("If a permanent dealt damage
// this way", Spikefield Hazard, Underworld Fires) — is marked. A player
// is not.
func ExileIfDealtDamageWouldDie(item *game.StackItem, permanents bool) RecipientThen {
	t := itemRecipient(item, exileIfDealtDamageWouldDie)
	if permanents {
		t.N = 1
	}
	return t
}

// exileIfDealtDamageWouldDie is ExileIfDealtDamageWouldDie's body: N 1 is
// "any permanent". An on-disk key.
var exileIfDealtDamageWouldDie = recipientTail("recipient/exile-if-dealt-damage-would-die", func(ctx *Context, target uuid.UUID, dealt int, o game.TailOutcome) error {
	if !dealtDamagePermanent(ctx.Game, target, dealt, o.Tail.N == 1) {
		return nil
	}
	return ExileIfItWouldDieThisTurn{Target: target}.Apply(ctx)
})

// CantBeRegeneratedIfDealtDamage is "A creature dealt damage this way
// can't be regenerated this turn" (Incinerate, Flamebreak, Jaya Ballard)
// as a continuation: a creature that was dealt more than 0 damage.
func CantBeRegeneratedIfDealtDamage(item *game.StackItem) RecipientThen {
	return itemRecipient(item, cantBeRegeneratedIfDealtDamage)
}

// cantBeRegeneratedIfDealtDamage is CantBeRegeneratedIfDealtDamage's
// body. An on-disk key.
var cantBeRegeneratedIfDealtDamage = recipientTail("recipient/cant-be-regenerated-if-dealt-damage", func(ctx *Context, target uuid.UUID, dealt int, _ game.TailOutcome) error {
	if !dealtDamagePermanent(ctx.Game, target, dealt, false) {
		return nil
	}
	return CantBeRegeneratedThisTurn{Target: target}.Apply(ctx)
})

// dealtDamagePermanent reports whether `target` was dealt damage and is
// a creature (or, with permanents set, any permanent) on the battlefield.
func dealtDamagePermanent(g *game.Game, target uuid.UUID, dealt int, permanents bool) bool {
	if dealt <= 0 {
		return false
	}
	z := g.FindCardZoneForEffect(target)
	if z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(target)
	return ok && (permanents || c.IsCreature())
}

// DealDamageThen is "~ deals `amount` damage to <target>", a player or a
// battlefield permanent, with `then` told the recipient and what it took
// once the damage has settled — after a CR 616 prompt, if one is asked.
// A target that is neither deals nothing and tells `then` nothing. The
// source is the resolving object's (ctx.Source()); a nil `then` is the
// fire-and-forget DealDamage.
func DealDamageThen(ctx *Context, target uuid.UUID, amount int, then RecipientThen) error {
	return DealDamageToEachThen(ctx, []uuid.UUID{target}, amount, then)
}

// DealDamageToEachThen is "~ deals `amount` damage to each of
// <targets>" with `then` told each recipient (Anger of the Gods,
// Flamebreak). The recipients are dealt their damage in order, each
// after the one before has settled (game.DealDamageEachEachKeyedForEffect).
func DealDamageToEachThen(ctx *Context, targets []uuid.UUID, amount int, then RecipientThen) error {
	if amount <= 0 || len(targets) == 0 {
		return nil
	}
	return ctx.Game.DealDamageEachEachKeyedForEffect(ctx.Source(), targets, amount, then, game.KeyedTail{})
}

// AllThen runs each continuation in order, for a card whose damage has
// more than one "dealt damage this way" rider (Runesword): each rider's
// Next is the one after it. Every rider must be a recipientTail.
func AllThen(thens ...RecipientThen) RecipientThen {
	var out RecipientThen
	for i := len(thens) - 1; i >= 0; i-- {
		t := thens[i]
		if t.IsZero() {
			continue
		}
		if !out.IsZero() {
			t = appendRecipient(t, out)
		}
		out = t
	}
	return out
}

// appendRecipient puts `next` at the end of t's rider chain.
func appendRecipient(t, next RecipientThen) RecipientThen {
	if t.Next == nil {
		t.Next = &next
		return t
	}
	inner := appendRecipient(*t.Next, next)
	t.Next = &inner
	return t
}

// --- the shared card bodies ------------------------------------------

// firstLegalTarget is the spell's one target if it is still legal
// (CR 608.2b): a player or a card, whichever the clause allows.
func firstLegalTarget(ctx *Context) (game.TargetRef, bool) {
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return game.TargetRef{}, false
	}
	return ts[0], true
}

// damageFirstTargetExileIfItDies is "~ deals N damage to target creature
// [or planeswalker]. If that creature [or planeswalker] would die this
// turn, exile it instead." (Lava Coil, Magma Spray, Scorching
// Dragonfire, Flame-Blessed Bolt and the rest), with the amount read as
// the spell resolves. The replacement is the spell's, not the damage's,
// so it is registered on a legal target whether or not the damage was
// dealt (the Disintegrate ruling).
func damageFirstTargetExileIfItDies(amount func(item *game.StackItem, ctx *Context) int) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: amount(item, ctx)}).Apply(ctx); err != nil {
			return err
		}
		return ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
	}
}

// fixedAmount is a damage amount printed as a number.
func fixedAmount(n int) func(*game.StackItem, *Context) int {
	return func(*game.StackItem, *Context) int { return n }
}

// xAmount is a damage amount of X. A call, not a value, so the XMatters
// guard sees the card read X.
func xAmount() func(*game.StackItem, *Context) int {
	return func(_ *game.StackItem, ctx *Context) int { return ctx.X() }
}

// damageAnyTargetExileIfDealtDies is "~ deals N damage to any target. If
// a creature [permanent] dealt damage this way would die this turn,
// exile it instead." (Demonfire, Annihilating Fire, Pillar of Flame,
// Yamabushi's Flame and the rest).
func damageAnyTargetExileIfDealtDies(amount func(item *game.StackItem, ctx *Context) int, permanents bool) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		t, ok := firstLegalTarget(ctx)
		if !ok {
			return nil
		}
		return DealDamageThen(ctx, t.ID, amount(item, ctx), ExileIfDealtDamageWouldDie(item, permanents))
	}
}

// damageThenIfCreatureNoRegenExileIfDies is "~ deals N damage to any
// target. If it's a creature, it can't be regenerated this turn, and if
// it would die this turn, exile it instead." (Disintegrate, Carbonize,
// a kicked Scorching Lava). Both riders are the spell's: a creature
// target is marked whether or not the damage was dealt (the Disintegrate
// ruling). A player or a noncreature permanent is only dealt damage.
func damageThenIfCreatureNoRegenExileIfDies(amount func(item *game.StackItem, ctx *Context) int) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		t, ok := firstLegalTarget(ctx)
		if !ok {
			return nil
		}
		if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount(item, ctx)}).Apply(ctx); err != nil {
			return err
		}
		return markCreatureNoRegenExileIfDies(ctx, t.ID)
	}
}

// markCreatureNoRegenExileIfDies is "If it's a creature, it can't be
// regenerated this turn, and if it would die this turn, exile it
// instead" on `id`, judged now.
func markCreatureNoRegenExileIfDies(ctx *Context, id uuid.UUID) error {
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if c, ok := ctx.Game.LookupCardForEffect(id); !ok || !c.IsCreature() {
		return nil
	}
	if err := (CantBeRegeneratedThisTurn{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
}

// damageAnyTargetNoRegenIfDealt is "~ deals N damage to any target. A
// creature dealt damage this way can't be regenerated this turn."
// (Incinerate, Jaya Ballard's second ability).
func damageAnyTargetNoRegenIfDealt(amount func(item *game.StackItem, ctx *Context) int) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		t, ok := firstLegalTarget(ctx)
		if !ok {
			return nil
		}
		return DealDamageThen(ctx, t.ID, amount(item, ctx), CantBeRegeneratedIfDealtDamage(item))
	}
}

// allCreaturesShrinkThenExileIfTheyDie is "All creatures get -N/-N until
// end of turn. If a creature [an opponent controls] would die this turn,
// exile it instead." (Flaying Tendrils, Malicious Malfunction, Malicious
// Eclipse). The shrink's set is locked as it begins (CR 611.2c); the
// replacement's is read live.
func allCreaturesShrinkThenExileIfTheyDie(n int, opponentsOnly bool) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		if err := (BoostUntilEOT{Match: Creature(), Power: n, Toughness: n}).Apply(ctx); err != nil {
			return err
		}
		return ExileIfCreaturesWouldDieThisTurn{OpponentsOnly: opponentsOnly}.Apply(ctx)
	}
}

// damageEachExileIfDealtDies is "~ deals N damage to each <predicate>.
// If a creature [permanent] dealt damage this way would die this turn,
// exile it instead." (Anger of the Gods, Crush the Weak, Yamabushi's
// Storm, Underworld Fires).
func damageEachExileIfDealtDies(ctx *Context, match CardPredicate, amount int, permanents bool) error {
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, match) {
		ids = append(ids, c.InstanceID)
	}
	return DealDamageToEachThen(ctx, ids, amount, ExileIfDealtDamageWouldDie(ctx.Item, permanents))
}
