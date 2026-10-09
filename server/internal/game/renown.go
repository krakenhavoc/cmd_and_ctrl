package game

// renown.go — CR 702.112, renown and the renowned designation (#2049,
// ADR 0071 amendment 2026-10-09). The sixth KEYWORD TRIGGER the engine
// derives from an object's ability list rather than from its catalog
// entry, and the second NUMBERED one after annihilator, whose shape it
// copies (annihilator.go).
//
//	702.112a Renown is a triggered ability. "Renown N" means "When this
//	         creature deals combat damage to a player, if it isn't
//	         renowned, put N +1/+1 counters on it and it becomes
//	         renowned."
//	702.112b Renowned is a designation that has no rules meaning other
//	         than to act as a marker that the renown ability and other
//	         spells and abilities can identify. Only permanents can be or
//	         become renowned. Once a permanent becomes renowned, it stays
//	         renowned until it leaves the battlefield. Renowned is
//	         neither an ability nor part of the permanent's copiable
//	         values.
//	702.112c If a creature has multiple instances of renown, each
//	         triggers separately. The first such ability to resolve will
//	         cause the creature to become renowned, and subsequent
//	         abilities will have no effect. (See rule 603.4)
//
// THE TOKEN. Renown carries a number, so it is stored the way
// annihilator is: KeywordRenown is the FAMILY key canonicalKeywords
// holds, the wire tokens are "renown N" (minted only by
// CanonicalRenownToken), and CanonicalKeywords refuses a bare "renown".
// Scryfall's keywords array says only "Renown", so the deck importer
// reads the number off the oracle line. It is CUMULATIVE (CR 702.112c):
// AppendKeywordAbility keeps a granted "renown 1" (Aragorn, Hornburg
// Hero) beside a printed one, and each is its own trigger. A token is
// what lets a creature whose only text is keywords and renown (Topan
// Freeblade, Citadel Castellan, …) need no card file at all.
//
// THE TRIGGER. "When this creature deals combat damage to a player":
// an EventDealDamage marked Combat, from this creature, to a player —
// never to a planeswalker, a battle or a creature — with a positive
// amount (prevented damage is not dealt, CR 615.1).
//
// THE INTERVENING IF (CR 603.4). "If it isn't renowned" is checked
// twice: as the damage is dealt (AppliesTo — a renowned creature does
// not trigger at all), and again as the trigger resolves
// (resolveRenownLocked — a creature that became renowned in between,
// by the first of two renown instances, gets nothing). That second
// check is all CR 702.112c needs: both instances trigger, and only the
// first to resolve does anything.
//
// THE RESOLUTION. Counters first, then the designation, as
// MonstrosityForEffect does: the N +1/+1 counters go through the CR 614
// counter window, placed by the ability's controller, so Hardened
// Scales and Doubling Season apply; the designation and
// EventBecameRenowned are the continuation, so a "whenever a creature
// becomes renowned" trigger never sees it before its counters. It
// becomes renowned even when the counters are replaced away: the
// instruction's second half is not conditional on the first. A source
// that has left the battlefield, or left and come back as a new object
// (CR 400.7), gets nothing, and the new object is not renowned.
//
// ORDERING (CR 603.3b). Two renown triggers on one creature do not
// commute when their numbers differ (only the first to resolve puts
// counters), so the label names N and the CR 603.3b prompt tells a
// renown 1 from a renown 2. Two with the same N are the same choice,
// and the existing same-source, same-label rule skips the prompt.

import (
	"strconv"

	"github.com/google/uuid"
)

// KeywordRenown is the bare word CR 702.112 spells with a number after
// it. Like KeywordAnnihilator it is the family key, not a token: the
// tokens are "renown N".
const KeywordRenown = "renown"

// maxRenownValue bounds what RenownValue accepts. Printed renown tops
// out at 6 (Outland Colossus), so a longer number is a malformed line
// and is refused, which errs weaker.
const maxRenownValue = 999

// RenownValue parses one ability token as CR 702.112's numbered
// keyword, reporting the N it carries. "renown" alone, "renown 0" and
// "renown two" answer false.
func RenownValue(token string) (int, bool) {
	return numberedKeywordValue(token, KeywordRenown, maxRenownValue)
}

// CanonicalRenownToken normalises one printed renown clause to the
// engine's wire form — "Renown 02" becomes "renown 2" — reporting
// whether it is one at all. The only thing that mints the token;
// CanonicalKeywords calls it.
func CanonicalRenownToken(s string) (string, bool) {
	n, ok := RenownValue(s)
	if !ok {
		return "", false
	}
	return KeywordRenown + " " + strconv.Itoa(n), true
}

// RenownAmounts is the N of every renown instance the card has, in
// ability-list order, through the walk keywordTriggersFor reads. A
// read for tests, the bot and the view.
func RenownAmounts(c *Card) []int {
	var out []int
	forEachAbilityToken(c, func(a string) bool {
		if n, ok := RenownValue(a); ok {
			out = append(out, n)
		}
		return true
	})
	return out
}

// renownLabel is the stack label of one renown N trigger. It names N,
// so two instances with different numbers on one creature are told
// apart by the CR 603.3b ordering prompt, and two with the same number
// are not.
func renownLabel(n int) string {
	what := "a +1/+1 counter"
	if n != 1 {
		what = strconv.Itoa(n) + " +1/+1 counters"
	}
	return "Renown " + strconv.Itoa(n) + " — if it isn't renowned, put " + what + " on it and it becomes renowned"
}

// renownTriggerFor is the one TriggeredAbility a single instance of
// "renown N" is.
func renownTriggerFor(n int) TriggeredAbility {
	label := renownLabel(n)
	return TriggeredAbility{
		Keyword: KeywordRenown,
		Watches: []EventKind{EventDealDamage},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
			// CR 603.4: "if it isn't renowned" — a renowned creature's
			// renown does not trigger.
			if source == nil || source.Renowned {
				return false
			}
			return renownDamageEvent(ev, source.InstanceID, g)
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			// Keyed directly, like annihilator: renown has no catalog
			// row, and a table with one waiting on the stack is still
			// a restore point.
			return NewKeyedTriggeredItem(source, label, renownGrowBody, EffectParams{Amount: n})
		},
	}
}

// renownDamageEvent is CR 702.112a's trigger event: combat damage, a
// positive amount of it, dealt by `creature` to a player.
//
// Caller must hold g.mu.
func renownDamageEvent(ev Event, creature uuid.UUID, g *Game) bool {
	if ev.Kind != EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != creature {
		return false
	}
	return g.playerByIDLocked(ev.Target) != nil
}

// renownGrowKey is the body every renown trigger names (ADR 0041 P9):
// an on-disk identity, never renamed or reused. An older binary refuses
// a restore point naming it with ErrUnknownEffectKey, which is the
// designed rollback case. Params.Amount is N.
const renownGrowKey = "renown/grow"

// renownGrowBody is the registered body's reference, built from the key
// and registered in init, as annihilator's is: the resolution places
// counters, which reaches the trigger harvest, which builds this item.
var renownGrowBody = BodyRef{key: renownGrowKey}

func init() {
	DelayedBody(renownGrowKey, func(g *Game, item *StackItem, p EffectParams) error {
		return g.resolveRenownLocked(item, p.Amount)
	})
}

// resolveRenownLocked is one renown trigger resolving: if its creature
// is still the object that triggered it and still isn't renowned (CR
// 603.4), put N +1/+1 counters on it, then it becomes renowned.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) resolveRenownLocked(item *StackItem, n int) error {
	if item == nil || g.AbilitySourceIsNewObjectForEffect(item) {
		return nil
	}
	self := findBattlefieldCard(g, item.SourceCardID)
	if self == nil || self.Renowned {
		return nil
	}
	return g.becomeRenownedLocked(item.Controller, self.InstanceID, n)
}

// becomeRenownedLocked puts n +1/+1 counters on the permanent through
// the CR 614 window and, as the placement's continuation, makes it
// renowned and emits EventBecameRenowned. The one writer of
// Card.Renowned outside the snapshot restore.
//
// Caller must hold g.mu in write mode.
func (g *Game) becomeRenownedLocked(placer, cardID uuid.UUID, n int) error {
	if n < 0 {
		n = 0
	}
	return g.AddCounterByThenForEffect(placer, cardID, CounterPlusOne, n, func(g *Game, _ int) error {
		// Re-find: the placement may have paused on a CR 616 prompt,
		// and the permanent pointer is not stable across an action.
		c := findBattlefieldCard(g, cardID)
		if c == nil || c.Renowned {
			return nil
		}
		c.Renowned = true
		g.EmitEvent(Event{
			Kind:   EventBecameRenowned,
			Actor:  c.Controller,
			Source: cardID,
			CardID: cardID,
			Target: cardID,
			Amount: n,
		})
		return nil
	})
}

// IsRenowned reports whether the named battlefield permanent is
// renowned — "if it's renowned" (Enshrouding Mist). False for anything
// not on the battlefield: CR 702.112b, only permanents can be renowned,
// and CR 400.7, a permanent that left is a new object.
//
// Caller must hold g.mu.
func (g *Game) IsRenowned(cardID uuid.UUID) bool {
	card := findBattlefieldCard(g, cardID)
	return card != nil && card.Renowned
}
