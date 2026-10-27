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
//     actually left, which a fire-and-forget loop never had. #866
//     extended that to exile and bounce: a verb with a `Then` runs it
//     from the sweep's continuation and is handed what LANDED, so
//     "for each card exiled this way" cannot be paid for a commander
//     that went to the command zone instead (CR 400.7).
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
	// isn't destroyed instead" — is dropped by the keyed batch's landed
	// list instead (#815). Optional; nil means the verb moves everything
	// it matched.
	narrow func(g *game.Game, ids []uuid.UUID) []uuid.UUID
	// move performs the batched zone change, fire-and-forget: nothing is
	// waiting on what moved, so a replacement's prompt lands its own card
	// later without holding the rest of the board up.
	move func(g *game.Game, ids []uuid.UUID) int
	// moveKeyed is the keyed door (#4253 slice 10a): the verb's batch
	// with its "for each permanent <moved> this way" clause as a
	// registered tail (MassTail), told the cards that actually LANDED
	// (#815, #866), because a leg of the batch can pause on a player
	// prompt and the set is not knowable until it is answered. Set, with
	// tail, when the verb's ThenTail is; otherwise move is used.
	moveKeyed func(g *game.Game, ids []uuid.UUID, tail game.KeyedTail) error
	tail      game.KeyedTail
}

func (m massEffect) apply(ctx *Context) error {
	swept := MatchingBattlefield(ctx, m.match)
	ids := make([]uuid.UUID, 0, len(swept))
	for _, c := range swept {
		ids = append(ids, c.InstanceID)
	}
	if m.narrow != nil {
		ids = m.narrow(ctx.Game, ids)
	}
	if m.moveKeyed != nil {
		// #4253 slice 10a: the clause is data on the batch, so a leg's
		// prompt is a restore point. The tail reads the batch's
		// pre-move copies itself (MassTail).
		return m.moveKeyed(ctx.Game, ids, m.tail)
	}
	m.move(ctx.Game, ids)
	return nil
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
	reader := ctx.textReader()
	var out []game.Card
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if match(ctx.Game, caster, game.ThroughTextOf(reader, c)) {
			out = append(out, c)
		}
	}
	return out
}

// DestroyAllMatching is "destroy all <predicate>" — Wrath of God,
// Damnation, Ritual of Soot, Planar Cleansing, an overloaded
// Vandalblast.
//
// "THEY CAN'T BE REGENERATED" IS ENFORCED, as of #667: set
// CantBeRegenerated and the sweep ignores regeneration shields
// (CR 701.19c) without spending them (CR 701.19c). It was cosmetic on
// every card that printed it for as long as the engine had no shield
// to ignore — the note Wrath of God carried from S14 to S40 — which
// is why every wipe in the catalog that prints the clause had to be
// revisited when the shield landed rather than only the new ones.
// Umbra armor (CR 702.89, effects.UmbraArmor) replaces a wipe's
// destruction of an enchanted creature like any other destruction.
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

	// ThenTail is the "for each permanent destroyed this way" clause, a
	// registered tail (MassTail, #4253 slice 10a) told the pre-move
	// copies of the permanents that were actually DESTROYED, so a table
	// waiting on a leg's prompt is a restore point. Optional.
	//
	// #815: it runs from a CONTINUATION, which means it may run an
	// action later than the sweep — a commander caught in the wipe
	// stops to answer a replacement's prompt, and the number is not
	// knowable until they do. Write it as a clause that acts on what it
	// is given, not as the next line of the card.
	ThenTail game.KeyedTail

	// CantBeRegenerated is the clause Damnation, Damn, Day of
	// Judgment, Decree of Pain, Winds of Rath and Shatterstorm print:
	// this destruction ignores regeneration shields (CR 701.19c), and
	// does not spend them (CR 701.19c). #667.
	CantBeRegenerated bool
}

func (d DestroyAllMatching) Apply(ctx *Context) error {
	opts := game.DestroyOptions{CantBeRegenerated: d.CantBeRegenerated}
	m := massEffect{
		match:  d.Match,
		narrow: (*game.Game).DestructibleForEffect,
	}
	if !d.ThenTail.IsZero() {
		m.moveKeyed = func(g *game.Game, ids []uuid.UUID, tail game.KeyedTail) error {
			return g.DestroyPermanentsKeyedForEffect(ids, tail, opts)
		}
		m.tail = d.ThenTail
		return m.apply(ctx)
	}
	m.move = func(g *game.Game, ids []uuid.UUID) int {
		return g.DestroyPermanentsForEffect(ids, opts)
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

	// ThenTail is the "for each card exiled this way" clause, a
	// registered tail (MassTail, #4253 slice 10a) told the pre-move
	// copies of the cards that actually reached EXILE. Optional.
	//
	// #866: it runs from a CONTINUATION, exactly as
	// DestroyAllMatching's does — a leg that stops to answer a
	// replacement's prompt makes the number unknowable until it does.
	// A commander is in it: it is exiled first and CR 903.9a asks its
	// owner afterwards (ADR 0115).
	ThenTail game.KeyedTail
}

func (e ExileAllMatching) Apply(ctx *Context) error {
	m := massEffect{match: e.Match}
	if !e.ThenTail.IsZero() {
		m.moveKeyed = (*game.Game).ExileCardsKeyedForEffect
		m.tail = e.ThenTail
		return m.apply(ctx)
	}
	m.move = func(g *game.Game, ids []uuid.UUID) int { return g.ExileCardsForEffect(ids) }
	return m.apply(ctx)
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

	// ThenTail is the "for each permanent returned this way" clause, a
	// registered tail (MassTail, #4253 slice 10a) told the permanents
	// returned this way, with the same continuation shape and the same
	// CR 400.7 reading as ExileAllMatching's. Optional.
	ThenTail game.KeyedTail
}

func (b BounceAllMatching) Apply(ctx *Context) error {
	m := massEffect{match: b.Match}
	if !b.ThenTail.IsZero() {
		m.moveKeyed = (*game.Game).BounceCardsToHandKeyedForEffect
		m.tail = b.ThenTail
		return m.apply(ctx)
	}
	m.move = func(g *game.Game, ids []uuid.UUID) int { return g.BounceCardsToHandForEffect(ids) }
	return m.apply(ctx)
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

	// ThenTail is the "for each permanent returned this way" clause, a
	// registered tail (#4253 slice 10a) told the cards returned this way
	// as TailOutcome.IDs, run from the bounce's continuation for the
	// reason BounceAllMatching.ThenTail is (#866). Optional.
	ThenTail game.KeyedTail
}

func (r ReturnAllToHand) Apply(ctx *Context) error {
	if !r.ThenTail.IsZero() {
		return ctx.Game.BounceCardsToHandKeyedForEffect(r.Cards, r.ThenTail)
	}
	ctx.Game.BounceCardsToHandForEffect(r.Cards)
	return nil
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
	return ctx.Game.DamageInstanceForEffect(func() error {
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
	})
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
