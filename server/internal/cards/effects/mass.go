package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mass.go — S23: the four mass-effect primitives.
//
// Every board wipe in the catalog before this file was a hand-rolled
// loop: snapshot the battlefield into a []uuid.UUID, then apply a
// single-target primitive per ID. Eleven cards carried some variant of
// the same eight lines and the same "snapshot first, the slice moves
// under you" comment. The primitives here are that loop, named, with
// the three things the hand-rolled version kept getting approximately
// right done in one place:
//
//  1. THE SNAPSHOT. The set is resolved once, before anything moves.
//     CR 608.2 — a resolving spell affects the objects that matched
//     when it looked, not ones that arrive mid-resolution.
//
//  2. SIMULTANEITY. The whole set leaves as ONE event, so a Blood
//     Artist caught in the wipe sees every creature that died with it
//     (CR 700.4 / 603.10). The per-ID loops did not, and the bug was
//     invisible in every single-target test. See
//     server/internal/game/simultaneous.go for the mechanism.
//
//  3. THE COUNT. "For each creature destroyed this way" (Fumigate,
//     Deadly Tempest, Bane of Progress) needs to know how many
//     actually left, which a fire-and-forget loop never had.
//
// The predicate is the whole vocabulary. "Destroy all creatures" is
// Match: Creature(); "destroy all creatures you don't control" is
// And(Creature(), OpponentControls()); "return all creatures except
// for Krakens, Leviathans, Octopuses and Serpents" is
// Except(Creature(), Subtype("Kraken"), …). Predicates compose out of
// targets.go, the same library the targeting clauses use, so a
// sweep's filter and a target's filter can never drift apart.

// massEffect is the shared body of the four primitives: resolve the
// predicate against the battlefield once, hand the IDs to a batched
// mover, then run the card's follow-up with the cards that were
// actually swept.
//
// Callers use the named primitives below rather than this directly —
// it exists so the four differ only in the verb.
type massEffect struct {
	match CardPredicate
	// narrow optionally removes permanents the VERB cannot move even
	// though the PREDICATE matched them — currently only
	// indestructible, which stops a destroy (CR 702.12b) and stops
	// nothing else. Applied before the move, so an indestructible
	// permanent is neither swept nor counted. A permanent the CR 614
	// window saves LATER — indestructible granted mid-window, "it
	// isn't destroyed instead" — is dropped by moveThen's landed list
	// instead (#815). Optional; nil means the verb moves everything it
	// matched.
	narrow func(g *game.Game, ids []uuid.UUID) []uuid.UUID
	// move performs the batched zone change and returns how many
	// cards actually moved. Set by a verb whose `then` is nil, or
	// whose move cannot pause; moveThen is the alternative.
	move func(g *game.Game, ids []uuid.UUID) int
	// moveThen is move for a verb that owes its `then` a CONTINUATION
	// rather than a return value, because a leg of the batch can pause
	// on a player prompt and the number is not knowable until it is
	// answered. It reports the cards that actually moved, not just how
	// many, so `swept` can be narrowed to them (#815).
	//
	// Exactly one of move / moveThen is set.
	moveThen func(g *game.Game, ids []uuid.UUID, then func(g *game.Game, moved []uuid.UUID) error) error
	// then is the card's "…for each permanent destroyed this way"
	// clause. Receives the pre-move copies and the count that actually
	// moved — of everything that was swept under `move`, and of
	// exactly what LANDED under `moveThen`, which knows.
	then func(ctx *Context, swept []game.Card, moved int) error
}

func (m massEffect) apply(ctx *Context) error {
	swept := MatchingBattlefield(ctx, m.match)
	ids := make([]uuid.UUID, 0, len(swept))
	for _, c := range swept {
		ids = append(ids, c.InstanceID)
	}
	if m.narrow != nil {
		ids = m.narrow(ctx.Game, ids)
		swept = keepCards(swept, ids)
	}
	if m.moveThen != nil {
		// The context is rebuilt inside the continuation from the live
		// *Game, the contract every continuation in the engine follows
		// — an undo restores this game's fields in place, so a captured
		// *Game would be the wrong one.
		item := ctx.Item
		return m.moveThen(ctx.Game, ids, func(g *game.Game, moved []uuid.UUID) error {
			if m.then == nil {
				return nil
			}
			// `swept` narrows to what actually moved, so the clause that
			// reads the CARDS and the clause that reads the COUNT agree
			// — a survivor must not be paid out for either way.
			return m.then(NewContext(g, item), keepCards(swept, moved), len(moved))
		})
	}
	moved := m.move(ctx.Game, ids)
	if m.then == nil {
		return nil
	}
	return m.then(ctx, swept, moved)
}

// keepCards narrows a swept set to the cards named in `keep`,
// preserving the swept order.
func keepCards(swept []game.Card, keep []uuid.UUID) []game.Card {
	want := make(map[uuid.UUID]bool, len(keep))
	for _, id := range keep {
		want[id] = true
	}
	kept := swept[:0:0]
	for _, c := range swept {
		if want[c.InstanceID] {
			kept = append(kept, c)
		}
	}
	return kept
}

// MatchingBattlefield resolves a predicate against the battlefield
// and returns COPIES of the matching cards, taken before anything
// moves. Cards rather than IDs because a follow-up clause usually
// wants to read the swept permanents ("each player loses life equal
// to the number of creatures THEY controlled") and by then they are
// in graveyards with their battlefield state cleared.
//
// Exported because a handful of cards need the set for something
// other than a sweep — Chandra's Ignition's "each other creature",
// Bane of Progress's counter count.
func MatchingBattlefield(ctx *Context, match CardPredicate) []game.Card {
	if match == nil {
		return nil
	}
	caster := ctx.Controller()
	var out []game.Card
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if match(ctx.Game, caster, c) {
			out = append(out, c)
		}
	}
	return out
}

// DestroyAllMatching is "destroy all <predicate>" — Wrath of God,
// Damnation, Ritual of Soot, Planar Cleansing, an overloaded
// Vandalblast.
//
// Regeneration and totem armor are not modelled anywhere in the
// engine, so "they can't be regenerated" is cosmetic on the cards
// that print it — the same note Wrath of God has carried since S14.
// The clause becomes load-bearing when regeneration lands, and this
// primitive is where it will be enforced.
//
// INDESTRUCTIBLE IS HONOURED, as of S30 (#470 / #446), and this
// comment has twice been the place the truth went stale — first
// claiming the engine had no such concept, then documenting the hole
// it left. The history is worth keeping because the failure mode is
// the recurring one: S25 (#77) shipped CR 702.12 in
// game/indestructible.go hours after S23 shipped this primitive, and
// gated only the single-target verb and the two damage branches of
// the SBA doomed pre-pass. The MASS path — g.DestroyPermanentsForEffect
// → destroyPermanentsLocked (game/simultaneous.go) — went straight to
// routeBattlefieldCardToOwnerGraveyardLocked and never asked, so
// every board wipe in the catalog killed an Avacyn for four sprints.
//
// The guard now lives at g.DestroyPermanentsForEffect, which filters
// through g.DestructibleForEffect before opening the simultaneity
// batch. This primitive applies the SAME filter to its own `swept`
// slice (massEffect.narrow), for the second half of the rule: a
// survivor must not be counted by "for each creature destroyed this
// way", and must not be reported to a Then clause that reads the
// cards rather than the count — Deadly Tempest's per-controller life
// loss is exactly that shape.
type DestroyAllMatching struct {
	// Match selects the permanents to destroy. Required; a nil
	// predicate sweeps nothing rather than everything, because
	// "destroy the entire battlefield" should never be the result of
	// a forgotten field.
	Match CardPredicate

	// Then is the "for each permanent destroyed this way" clause.
	// `swept` holds the pre-move copies of the permanents that were
	// actually DESTROYED and `destroyed` is how many of them there
	// were, so the two always describe the same set. Optional.
	//
	// #815: it runs from a CONTINUATION, which means it may run an
	// action later than the sweep — a commander caught in the wipe
	// stops to answer CR 903.9, and the number is not knowable until
	// they do. Write it as a clause that acts on what it is given, not
	// as the next line of the card.
	Then func(ctx *Context, swept []game.Card, destroyed int) error
}

func (d DestroyAllMatching) Apply(ctx *Context) error {
	m := massEffect{
		match:  d.Match,
		narrow: (*game.Game).DestructibleForEffect,
		then:   d.Then,
	}
	if d.Then == nil {
		// Nothing is waiting on the count, so the sweep stays
		// fire-and-forget: every leg is destroyed on this line, and a
		// commander's CR 903.9 prompt lands its own card later without
		// holding the rest of the board up.
		m.move = func(g *game.Game, ids []uuid.UUID) int { return g.DestroyPermanentsForEffect(ids) }
	} else {
		m.moveThen = func(g *game.Game, ids []uuid.UUID, then func(*game.Game, []uuid.UUID) error) error {
			return g.DestroyPermanentsThenForEffect(ids, then)
		}
	}
	return m.apply(ctx)
}

// ExileAllMatching is "exile all <predicate>" — Merciless Eviction,
// Farewell.
//
// The reason a deck plays this over DestroyAllMatching is entirely
// about what does NOT happen: no dies-triggers, no graveyard
// recursion, and no indestructible — exile is not destruction, so
// CR 702.12b never gets a say and there is deliberately no `narrow`
// hook here (regeneration, once it exists, is the same story).
// Against the catalog's aristocrats package
// that is the difference between a wipe that drains the table and
// one that doesn't.
type ExileAllMatching struct {
	Match CardPredicate
	Then  func(ctx *Context, swept []game.Card, exiled int) error
}

func (e ExileAllMatching) Apply(ctx *Context) error {
	return massEffect{
		match: e.Match,
		move:  func(g *game.Game, ids []uuid.UUID) int { return g.ExileCardsForEffect(ids) },
		then:  e.Then,
	}.apply(ctx)
}

// BounceAllMatching is "return all <predicate> to their owners'
// hands" — Evacuation, River's Rebuke, Whelming Wave, an overloaded
// Cyclonic Rift.
//
// Bounce is the sweep that beats indestructible, tokens (they cease
// to exist) and graveyard recursion at once, and the only one whose
// victims come back — which is why the blue ones cost so much more
// than Wrath of God does.
type BounceAllMatching struct {
	Match CardPredicate
	Then  func(ctx *Context, swept []game.Card, bounced int) error
}

func (b BounceAllMatching) Apply(ctx *Context) error {
	return massEffect{
		match: b.Match,
		move:  func(g *game.Game, ids []uuid.UUID) int { return g.BounceCardsToHandForEffect(ids) },
		then:  b.Then,
	}.apply(ctx)
}

// ReturnAllToHand is BounceAllMatching's explicit-set sibling: return
// THIS set of cards to their owners' hands, as one simultaneous
// event.
//
// It exists because not every mass bounce is expressible as a
// battlefield predicate. Aetherize returns "all attacking creatures",
// and attacking-ness lives in the combat state rather than on the
// card, so the set is computed by the caller and handed over. The
// batching and the ordering guarantees are identical.
type ReturnAllToHand struct {
	// Cards is the set to return, in the order the caller wants them
	// processed. IDs that are not on the battlefield are skipped.
	Cards []uuid.UUID

	Then func(ctx *Context, bounced int) error
}

func (r ReturnAllToHand) Apply(ctx *Context) error {
	bounced := ctx.Game.BounceCardsToHandForEffect(r.Cards)
	if r.Then == nil {
		return nil
	}
	return r.Then(ctx, bounced)
}

// --- helpers the mass cards share --------------------------------

// damageEachMatching is "<source> deals N damage to each
// <predicate>" — Pyroclasm, Blasphemous Act, Brotherhood's End, Star
// of Extinction.
//
// Deliberately NOT one of the four primitives, because mass damage is
// not a mass zone change. Nothing leaves the battlefield here: each
// creature takes its damage, and the ones that took lethal die later,
// when the state-based-action sweep runs at the next boundary. That
// sweep is itself batched (see game/simultaneous.go), so the deaths
// end up simultaneous anyway — but they are the SBA's deaths, not
// this effect's, which is why an indestructible creature (S25) or a
// damage prevention shield (S30) changes the outcome here and a
// DestroyAllMatching would not have let it.
func damageEachMatching(ctx *Context, match CardPredicate, amount int) error {
	if amount <= 0 {
		return nil
	}
	for _, c := range MatchingBattlefield(ctx, match) {
		if err := (DealDamage{
			Source: ctx.Source(),
			Target: c.InstanceID,
			Amount: amount,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// controllersOf tallies how many of `cards` each player controlled,
// for the "each player loses life equal to the number of creatures
// they controlled that were destroyed this way" family (Deadly
// Tempest). Reads the PRE-MOVE copies, which is the only place that
// information still exists once the sweep has run.
func controllersOf(cards []game.Card) map[uuid.UUID]int {
	out := make(map[uuid.UUID]int, len(cards))
	for _, c := range cards {
		out[c.Controller]++
	}
	return out
}
