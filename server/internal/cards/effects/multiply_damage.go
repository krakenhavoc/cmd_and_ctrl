package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// multiply_damage.go — the card-facing half of ADR 0108 §3 (#1890):
// "it deals double (triple) that damage instead" created by a resolving
// spell or ability. The engine half, game/multiply_damage.go, says why it
// is a replacement and not a prevention effect (CR 614.1a), why "a source
// you control" is read as the damage is dealt (CR 611.2c) and why "the
// next time" is one instance of damage (CR 615.8).
//
// Append-only, mechanic-named. Writing a card:
//
//	MultiplyDamage{Factor: 2, Sources: game.DamageSourcesYours}
//	  — Insult: "If a source you control would deal damage this turn"
//	MultiplyDamage{Factor: 3, Sources: game.DamageSourcesYours,
//	    Recipients: game.DamageRecipientsOpponentsAndTheirPermanents}
//	  — Isengard Unleashed
//	MultiplyDamage{Factor: 2, From: attacker, Next: true, CombatOnly: true}
//	  — Impulsive Maneuvers' "the next time that creature would deal
//	    combat damage this turn"

// MultiplyDamage registers a damage multiplier for the rest of the turn
// (or until a player's next turn). From names one source; with none,
// Sources says which (any source when it too is empty).
type MultiplyDamage struct {
	// Factor is 2 for "double", 3 for "triple".
	Factor int

	// From is one named source — a target, an attacking creature — pinned
	// as the object it is now (CR 400.7). FromRef is the same for a
	// source already pinned (a chosen source, CR 609.7a), with FromZone
	// the zone it was chosen in.
	From     uuid.UUID
	FromRef  game.ObjectRef
	FromZone game.ZoneKind

	// Sources is "a source you control" / "a creature" when no one source
	// is named.
	Sources game.DamageSources

	// Recipients is "to what"; Player is the player
	// DamageRecipientsPlayerAndTheirPermanents names.
	Recipients game.DamageRecipients
	Player     uuid.UUID

	// CombatOnly is "would deal combat damage"; Next is "the next time".
	CombatOnly bool
	Next       bool

	// UntilYourNextTurn is "until your next turn" — the resolving
	// effect's controller's (Lightning, Army of One) — instead of "this
	// turn".
	UntilYourNextTurn bool

	// Label is the record's; it defaults to the card's name.
	Label string
}

// Apply registers the record. A named source in no zone registers
// nothing (CR 608.2b's "it's gone").
func (m MultiplyDamage) Apply(ctx *Context) error {
	g := ctx.Game
	d := game.DamageMultiplier{
		EffectSource: ctx.Source(),
		Controller:   ctx.Controller(),
		Factor:       m.Factor,
		Sources:      m.Sources,
		Recipients:   m.Recipients,
		Player:       m.Player,
		CombatOnly:   m.CombatOnly,
		Next:         m.Next,
		Label:        m.Label,
	}
	if d.Label == "" {
		d.Label = shieldSourceName(ctx)
	}
	if m.UntilYourNextTurn {
		d.UntilNextTurnOf = ctx.Controller()
	}
	switch {
	case m.FromRef.ID != uuid.Nil:
		d.Source, d.SourceZone = m.FromRef, m.FromZone
	case m.From != uuid.Nil:
		ref, zone, ok := g.DamageSourceRefLocked(m.From)
		if !ok {
			return nil
		}
		d.Source, d.SourceZone = ref, zone
	}
	g.MultiplyDamageForEffect(d)
	return nil
}

// NextTimeFlip is the shape Desperate Gambit and Impulsive Maneuvers
// share: "flip a coin. If you win the flip, the next time <source> would
// deal [combat] damage this turn, it deals double that damage instead.
// If you lose the flip, the next time it would deal [combat] damage this
// turn, prevent that damage." The flip is the resolving effect's
// controller's to call and to win or lose (CR 705.2); the source is
// pinned as it is now (CR 400.7). A zero ref is a source that could not
// be named: the coin is still flipped, and nothing else happens.
type NextTimeFlip struct {
	Source     game.ObjectRef
	SourceZone game.ZoneKind
	CombatOnly bool
	Question   string
}

// nextTimeFlipBy is the resolving item a NextTimeFlip belongs to, as
// plain values: a continuation (a choice answered later) rebuilds its
// context from these, never from the item it began with.
type nextTimeFlipBy struct {
	controller, source, self uuid.UUID
	kind                     game.StackItemKind
	name                     string
}

// Apply queues the coin call for the resolving item.
func (f NextTimeFlip) Apply(ctx *Context) error {
	f.queue(ctx.Game, nextTimeFlipBy{
		controller: ctx.Controller(), source: ctx.Source(), self: ctx.Item.ID,
		kind: ctx.Item.Kind, name: shieldSourceName(ctx),
	})
	return nil
}

// queue queues the coin call. Both halves are built in the flip's
// continuation from plain values.
func (f NextTimeFlip) queue(g *game.Game, by nextTimeFlipBy) {
	controller, source, name := by.controller, by.source, by.name
	question := f.Question
	if question == "" {
		question = name + " — call the coin flip"
	}
	combat := 0
	if f.CombatOnly {
		combat = 1
	}
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  controller,
		Source:   source,
		Question: question,
		// Keyed (#4253 slice 10b): the resolving item and the pinned
		// source ride the call as values.
		ThenKey: nextTimeFlipped,
		Carry:   []uuid.UUID{by.self, f.Source.ID},
		Nums:    []int{f.Source.Epoch, combat},
		Words:   []string{name, string(f.SourceZone), string(by.kind)},
	})
}

// nextTimeFlipped is a NextTimeFlip's settled call: a won flip doubles the
// source's next damage, a lost one prevents it. An on-disk key: never
// renamed, never reused.
var nextTimeFlipped = game.RegisterCoinFlipThen("coin-flip/next-time-double-or-prevent", func(g *game.Game, r game.CoinFlipped) error {
	src := game.ObjectRef{ID: carriedID(r.Carry, 1), Epoch: carriedNum(r.Nums, 0)}
	if src.ID == uuid.Nil {
		return nil
	}
	name, zone, combat := carriedWord(r.Words, 0), game.ZoneKind(carriedWord(r.Words, 1)), carriedNum(r.Nums, 1) == 1
	controller, source := r.Flipper, r.Source
	c := NewContext(g, &game.StackItem{
		ID: carriedID(r.Carry, 0), Kind: game.StackItemKind(carriedWord(r.Words, 2)),
		Controller: controller, Owner: controller, SourceCardID: source,
	})
	if wonFirstFlip(r) {
		return MultiplyDamage{Factor: 2, FromRef: src, FromZone: zone,
			CombatOnly: combat, Next: true, Label: name + " — double the next damage"}.Apply(c)
	}
	g.PreventNextDamageFromSourceForEffect(game.NextDamageShield{
		EffectSource: source,
		Controller:   controller,
		Source:       src,
		SourceZone:   zone,
		CombatOnly:   combat,
		Label:        name + " — prevent the next damage",
	})
	return nil
})
