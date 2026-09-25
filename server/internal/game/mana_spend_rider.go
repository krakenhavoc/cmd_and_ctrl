package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// mana_spend_rider.go — #1547, ADR 0040 amendment 2026-09-24: mana that
// does something WHEN IT IS SPENT.
//
//	Cavern of Souls        "…and that spell can't be countered."
//	Hall of the Bandit Lord "If that mana is spent on a creature spell,
//	                        it gains haste."
//	Pyromancer's Goggles   "When that mana is spent to cast a red
//	                        instant or sorcery spell, copy that spell…"
//	Biophagus              "If this mana is spent to cast a creature
//	                        spell, that creature enters with an
//	                        additional +1/+1 counter on it."
//
// Every one of these is a fact about ONE mana's journey: made by this
// source, then spent on that spell. The pool already carried the first
// half — a ManaToken knows its Source (S15), its spend restrictions
// (#352) and what the source was (#1212) — and #761 made the second half
// a record: StackItem.Paid.Mana is the tokens that actually paid. What
// was missing was a way for the token to CARRY an instruction from the
// first moment to the second.
//
// # A rider is DATA on the token
//
// The same argument ADR 0040 §2 made for restrictions, for the same three
// reasons: a token with a rider sits in a pool across an undo boundary,
// rides the snapshot to disk, and is read back after the source that made
// it has gone. So ManaSpendRider is a kind, a filter in the restriction
// vocabulary, and a few plain fields. The one part that has to be code —
// what a "when that mana is spent" TRIGGER does — is named by a string
// key into a process-lifetime registry the catalog fills at init, exactly
// as a catalog hook is named by an oracle ID. The key survives the
// snapshot; the closure is rebuilt by the new binary.
//
// # It fires at the ONE place a payment becomes a stack object
//
// A rider fires when the token that carries it is spent paying for a
// spell (castSpellLocked) or an activated ability (activateAbilityLocked)
// and its filter admits the object — both sites call
// applyManaSpendRidersLocked right after the item is registered, with the
// SAME spend context the payment itself was solved under. Manual payment
// and the auto-tapper arrive here identically, because the auto-tapper
// only ever puts mana in the pool: the spend is the ordinary strict-mode
// spend either way, and the tokens it mints carry the ability's riders
// through the same produceManaLocked body.
//
// Firing stamps Applied on the rider's copy in item.Paid.Mana. That copy
// is the record every reader consults:
//
//   - "can't be countered" — spellCantBeCounteredLocked, the one gate
//     every counter verb asks (#1318's exit primitive sits behind it).
//   - "enters with an additional counter" — applyCastEntryCountersLocked,
//     the one seeding site of the entry event.
//   - "gains haste" — the layer gather below, reading the permanent's
//     Card.Provenance.Mana, which is item.Paid.Mana carried across the
//     entry (CR 400.7d, #1212).
//
// A trigger rider is the one kind that ACTS at the spend: it goes on the
// PendingTriggers queue there, and runStateChecksLocked at the end of the
// cast places it above the spell — CR 603.3, the next time a player would
// receive priority, which is right after CR 601.2i finishes the cast.
//
// # What does NOT fire a rider, deliberately
//
//   - An OnPaper payment (permissive mode, ForceCast). The engine spent
//     nothing, so no token left the pool and there is nothing to ride.
//     Weaker than printed — ADR 0068 §3's posture for every reader of the
//     spend record — and the cards say so in their caveats.
//   - A mana ability's own mana cost (a Signet's {1}). There is no stack
//     object to carry an uncounterable flag or to be copied, and no
//     printed rider asks about one.
//   - A pay-unless / ward-style tax (payCostLocked). It spends under the
//     zero context, which no filter admits.
//   - A COPY of a spell. CR 707.10: the copy was not cast, so no mana was
//     spent on it; copiedPaidCost carries no tokens. A Goggles copy is not
//     uncounterable and does not trigger the Goggles again.

// ManaRiderKind names what a rider does once it fires.
type ManaRiderKind string

const (
	// ManaRiderCantBeCountered — "that spell can't be countered"
	// (Cavern of Souls, Delighted Halfling, Boseiju, Who Shelters All).
	// A property of the stack item, read by spellCantBeCounteredLocked.
	// Meaningful on a spell only.
	ManaRiderCantBeCountered ManaRiderKind = "cant_be_countered"

	// ManaRiderHaste — "it gains haste" (Hall of the Bandit Lord): the
	// permanent the spell becomes has haste for as long as it remains
	// that object. Read off Card.Provenance by the layer gather.
	ManaRiderHaste ManaRiderKind = "haste"

	// ManaRiderEntersWithCounters — "that creature enters with an
	// additional +1/+1 counter on it" (Biophagus). CounterKind and
	// Counters say which and how many; seeded onto the entry event with
	// the card's own CR 614.1c clauses, so Doubling Season sees them.
	ManaRiderEntersWithCounters ManaRiderKind = "enters_with_counters"

	// ManaRiderTrigger — "when that mana is spent to cast …, <effect>"
	// (Pyromancer's Goggles, Scaled Nurturer, Path of Ancestry). A
	// triggered ability (CR 603.2) whose effect is registered under
	// Trigger.
	ManaRiderTrigger ManaRiderKind = "trigger"
)

// ManaSpendRider is one "when this mana is spent" instruction riding on a
// ManaToken. Pure data: see the file comment.
type ManaSpendRider struct {
	Kind ManaRiderKind `json:"kind"`

	// When is the spend filter, in the restriction vocabulary
	// (mana_restriction.go) and matched by the same ManaSpendContext
	// method: every tag must hold. Empty admits any spend. An unknown tag
	// admits nothing — a rider that cannot say what it applies to does
	// not fire, which is the weaker direction.
	//
	// Distinct from ManaToken.Restrictions, which decide whether the
	// token may pay AT ALL. Hall of the Bandit Lord's {C} pays for
	// anything and hastes only a creature; Cavern's coloured mana pays
	// only for a creature of the chosen type, and every such spell is
	// uncounterable.
	When []string `json:"when,omitempty"`

	// CounterKind and Counters are ManaRiderEntersWithCounters' payload.
	CounterKind string `json:"counterKind,omitempty"`
	Counters    int    `json:"counters,omitempty"`

	// Trigger is ManaRiderTrigger's registry key (RegisterManaSpendTrigger).
	Trigger string `json:"trigger,omitempty"`

	// Production identifies the ONE production that minted the token —
	// every token one call to produceManaLocked puts in the pool shares
	// it. "When THAT mana is spent" is a statement about the mana one
	// activation made: a Mana Reflection that doubles a Goggles' {R} is
	// still one Goggles activation and still one copy, and a Biophagus
	// {G} doubled is still one extra counter. So riders are deduplicated
	// on it when they fire. Stamped by produceManaLocked; zero on a
	// declaration.
	Production uuid.UUID `json:"production,omitempty"`

	// Applied is set on the copy in StackItem.Paid.Mana when the filter
	// admitted the object the token paid for. It is the record the
	// readers consult; a rider that rode a token onto an object outside
	// its filter stays false and does nothing (Hall mana spent on a
	// sorcery).
	Applied bool `json:"applied,omitempty"`
}

// ManaSpendTrigger is the code half of a ManaRiderTrigger: what the
// "when that mana is spent" ability does. Registered once per card at
// init (RegisterManaSpendTrigger) and looked up by key when the rider
// fires, so the token itself never holds a closure.
type ManaSpendTrigger struct {
	// Label is the triggered ability's stack label.
	Label string

	// Condition narrows the filter where the tag vocabulary cannot say
	// it — Path of Ancestry's "a creature spell that shares a creature
	// type with your commander". Evaluated at the spend, against the
	// object paid for; nil means the When tags are the whole filter.
	//
	// Runs under g.mu held for write; read-only.
	Condition func(g *Game, controller uuid.UUID, paidFor Card) bool

	// Effect is the ability's effect. The item it receives has
	// Controller = the player who spent the mana, SourceCardID = the
	// permanent that made it (which may be gone), and Payload[0] = the
	// spell or ability the mana paid for, as a TargetCard ref — "copy
	// THAT SPELL" is a Payload read, not a target (it does not target,
	// CR 115.10), so a spell countered in response is still named.
	//
	// Same contract as every triggered effect: read everything off the
	// item and the *Game handed in; capture nothing.
	Effect func(g *Game, item *StackItem) error
}

var manaSpendTriggers = struct {
	sync.RWMutex
	byKey map[string]ManaSpendTrigger
}{byKey: map[string]ManaSpendTrigger{}}

// RegisterManaSpendTrigger registers the code half of a "when that mana
// is spent" rider under `key`. Called from catalog init; panics on an
// empty key, a nil Effect or a duplicate, because each of those is a
// card-file bug that would otherwise surface as a rider that silently
// never fires.
func RegisterManaSpendTrigger(key string, t ManaSpendTrigger) {
	if key == "" {
		panic("game: RegisterManaSpendTrigger with an empty key")
	}
	if t.Effect == nil {
		panic(fmt.Sprintf("game: mana spend trigger %q has no Effect", key))
	}
	manaSpendTriggers.Lock()
	defer manaSpendTriggers.Unlock()
	if _, dup := manaSpendTriggers.byKey[key]; dup {
		panic(fmt.Sprintf("game: mana spend trigger %q registered twice", key))
	}
	manaSpendTriggers.byKey[key] = t
}

// ManaSpendTriggerFor returns the registered trigger for `key`.
func ManaSpendTriggerFor(key string) (ManaSpendTrigger, bool) {
	manaSpendTriggers.RLock()
	defer manaSpendTriggers.RUnlock()
	t, ok := manaSpendTriggers.byKey[key]
	return t, ok
}

// manaRiderDispatchBody is "mana-rider/dispatch" (ADR 0041 P9, #1497,
// tier 4): the one tier-2 body every mana-spend-rider trigger's item
// carries, re-deriving the per-card ManaSpendTrigger from Params.Name —
// the same manaSpendTriggers key ManaSpendTriggerFor already reads, and
// already an on-disk identity (ManaRider.Trigger rides the snapshot on
// the paying token). An unknown key is the same "weaker than printed"
// posture applyManaSpendRidersLocked already takes for one, rather than
// a refusal: the rider registry is its own append-only vocabulary, not
// the tier-2 ledger's.
var manaRiderDispatchBody = DelayedBody("mana-rider/dispatch", func(g *Game, item *StackItem, p EffectParams) error {
	t, ok := ManaSpendTriggerFor(p.Name)
	if !ok {
		return nil
	}
	return t.Effect(g, item)
})

// copyManaRiders returns a fresh backing array for a rider list (and for
// each rider's filter), or nil for an empty one — copyRestrictions' twin
// and for its reason: the catalog's slice is process-lifetime and shared
// by every instance of the card, while a token is game state.
func copyManaRiders(rs []ManaSpendRider) []ManaSpendRider {
	if len(rs) == 0 {
		return nil
	}
	out := make([]ManaSpendRider, len(rs))
	for i, r := range rs {
		out[i] = r
		out[i].When = copyRestrictions(r.When)
	}
	return out
}

// stampProduction returns the ability's riders as one production's:
// fresh backing arrays, and one Production id shared by all of them.
// Nil in, nil out, and no id is minted for nothing.
func stampProduction(rs []ManaSpendRider) []ManaSpendRider {
	out := copyManaRiders(rs)
	if len(out) == 0 {
		return nil
	}
	id := uuid.New()
	for i := range out {
		out[i].Production = id
	}
	return out
}

// clone deep-copies a token: its restriction list and its riders. Every
// place a token is copied into another home — the pool across undo, the
// payment record, the permanent's provenance, the snapshot — goes through
// here, so a new slice field on ManaToken is one edit, not six.
func (t ManaToken) clone() ManaToken {
	out := t
	if len(t.Restrictions) > 0 {
		out.Restrictions = append([]string(nil), t.Restrictions...)
	}
	out.Riders = copyManaRiders(t.Riders)
	return out
}

// cloneManaTokens deep-copies a token slice, nil for an empty one.
func cloneManaTokens(ts []ManaToken) []ManaToken {
	if len(ts) == 0 {
		return nil
	}
	out := make([]ManaToken, len(ts))
	for i, t := range ts {
		out[i] = t.clone()
	}
	return out
}

// applyManaSpendRidersLocked fires every rider on the tokens that paid
// for `item`, against `ctx` — the spend context the payment was solved
// under — and `paidFor`, the object itself (the spell's card, or the
// ability's source).
//
// Stamps Applied on each admitted rider's copy in item.Paid.Mana and
// queues one trigger per admitted trigger rider per production. Does
// nothing for an OnPaper payment, which spent no tokens.
//
// Caller must hold g.mu in write mode, with `item` already registered
// in StackMeta.
func (g *Game) applyManaSpendRidersLocked(item *StackItem, ctx ManaSpendContext, paidFor Card) {
	if item == nil || len(item.Paid.Mana) == 0 {
		return
	}
	fired := map[uuid.UUID]map[string]bool{}
	for ti := range item.Paid.Mana {
		tok := &item.Paid.Mana[ti]
		for ri := range tok.Riders {
			r := &tok.Riders[ri]
			if !ctx.allows(r.When) {
				continue
			}
			if r.Kind == ManaRiderTrigger {
				t, ok := ManaSpendTriggerFor(r.Trigger)
				if !ok {
					// A key this binary does not know — a snapshot from
					// a build that had the card. Weaker than printed.
					continue
				}
				if t.Condition != nil && !t.Condition(g, item.Controller, paidFor) {
					continue
				}
				// Deduplicate on the production: one activation's mana
				// is one "that mana", however many tokens it became.
				seen := fired[r.Production]
				if seen == nil {
					seen = map[string]bool{}
					fired[r.Production] = seen
				}
				if r.Production != uuid.Nil && seen[r.Trigger] {
					r.Applied = true
					continue
				}
				seen[r.Trigger] = true
				r.Applied = true
				// ADR 0041 P9 (#1497, tier 4): the four WhenManaSpent
				// cards have no catalog row for this item, so it is
				// keyed directly under "mana-rider/dispatch" — one
				// body, re-looking up t.Effect by Params.Name (the
				// SAME registry key, ManaSpendTriggerFor already reads)
				// rather than a per-card key, since the registry
				// already IS the append-only identity a rider's
				// ManaRider.Trigger field carries in the snapshot.
				params := EffectParams{Name: r.Trigger}
				g.queueHarvestedTriggerLocked(&StackItem{
					Kind:         StackItemTriggered,
					Controller:   item.Controller,
					Owner:        item.Controller,
					SourceCardID: tok.Source,
					Label:        t.Label,
					Payload:      []TargetRef{{Kind: TargetCard, ID: item.ID}},
					Body:         manaRiderDispatchBody.Key(),
					Params:       params,
					Effect:       bodyEffect(manaRiderDispatchBody.Key(), params),
				})
				continue
			}
			r.Applied = true
		}
	}
}

// ridersApplied reports whether any token that paid for `item` carries an
// applied rider of `kind`.
func ridersApplied(mana []ManaToken, kind ManaRiderKind) bool {
	for _, t := range mana {
		for _, r := range t.Riders {
			if r.Applied && r.Kind == kind {
				return true
			}
		}
	}
	return false
}

// SpellCantBeCounteredByMana reports whether the mana that paid for this
// stack item made it uncounterable — Cavern of Souls, Delighted Halfling,
// Boseiju. Exported for the view and for tests; the engine's own gate is
// spellCantBeCounteredLocked, which asks this.
func (item *StackItem) SpellCantBeCounteredByMana() bool {
	return item != nil && item.Kind == StackItemSpell && ridersApplied(item.Paid.Mana, ManaRiderCantBeCountered)
}

// riderEntryCounters folds the applied ManaRiderEntersWithCounters riders
// on `item`'s payment into the entry event — one grant per production,
// so a doubled Biophagus {G} is still one additional counter.
func riderEntryCounters(ev *ReplacementEvent, item *StackItem) {
	if ev == nil || item == nil {
		return
	}
	seen := map[uuid.UUID]bool{}
	for _, t := range item.Paid.Mana {
		for _, r := range t.Riders {
			if !r.Applied || r.Kind != ManaRiderEntersWithCounters || r.CounterKind == "" || r.Counters <= 0 {
				continue
			}
			if r.Production != uuid.Nil {
				if seen[r.Production] {
					continue
				}
				seen[r.Production] = true
			}
			ev.AddCounterAtETB(r.CounterKind, r.Counters)
		}
	}
}

// manaRiderContinuousEffectsLocked is the layer-6 half of
// ManaRiderHaste: one self-only keyword grant per permanent whose
// provenance says an applied haste rider paid for the spell it was.
//
// Gathered like any other continuous effect rather than folded into the
// printed baseline, so "loses all abilities" with a later timestamp still
// takes the haste away (CR 613.7). Built fresh each pass from data on the
// card, which is what lets it survive the snapshot: nothing is
// registered, so there is no closure for the continuation census to
// count. The timestamp is the permanent's own — the effect began when the
// spell became it.
//
// Not `live`: the effect was created by a mana ability that has long
// since resolved, not by a static ability of the permanent, so removing
// the permanent's abilities must not silence it as a SOURCE (it strips
// the keyword in layer 6 instead, which is the rule).
//
// Caller must hold g.mu in write mode (the layer pass).
func (g *Game) manaRiderContinuousEffectsLocked() []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if !ridersApplied(src.Provenance.Mana, ManaRiderHaste) {
			continue
		}
		id := src.InstanceID
		out = append(out, staticContinuousEffect{
			ability: StaticAbility{
				Layer: Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
					return target != nil && target.InstanceID == id
				},
				Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
					if !containsKeyword(ch.Abilities, "haste") {
						ch.Abilities = append(ch.Abilities, "haste")
					}
				},
			},
			source:    src,
			timestamp: src.layerTimestamp(),
		})
	}
	return out
}
