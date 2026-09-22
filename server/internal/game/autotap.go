package game

import (
	"sort"

	"github.com/google/uuid"
)

// autotap.go is the S15 sub-PR 4 backtracking auto-tapper. Public
// API: `Game.AutoTapForCost(controller, cost, xValue)` returns a
// plan (list of permanent IDs to tap) that satisfies the parsed
// cost, or (nil, false) when no plan fits within the budget.
// `AutoTapForCostExcluding` accepts a set of "lock-tap" excluded
// sources the player has pre-pinned for a different purpose; the
// tapper solves the remaining cost from the surviving pool.
//
// Algorithm: greedy-with-backtracking, restriction-first heuristic.
// Each available source contributes a list of `ProducedManaEntry`
// slots — Sol Ring is `[{C}, {C}]`, Birds of Paradise is
// `[{W,U,B,R,G}]`, Arcane Signet (Bant) is `[{W,U,G}]`. For each
// colored requirement in `cost.Required`, the solver finds the
// most-restrictive un-used source whose first matching slot can
// cover it; on failure it backtracks to the previous requirement
// and tries the next-best source. After all colored requirements
// land, the remaining slots across used sources are tallied
// against `cost.Generic + cost.XSlots*xValue`; if short, the
// solver recruits additional unused sources (preferring colorless
// producers) until the budget is met or the source pool runs out.
//
// Budget cap: 10k node expansions. A pathological manabase
// (12+ duals + filter lands + tri-lands) can blow up the search
// tree; the cap returns (nil, false) cleanly so the client falls
// back to manual tapping. Real Commander manabases (38 lands +
// rocks) resolve in microseconds.
//
// What this sub-PR does NOT do: filter-land sub-payment (S17),
// Cavern of Souls tribe-locking (later sprint), phyrexian self-
// pay life cost (S17), hybrid-color preferences (greedy picks
// the first matching half), `{X}` mid-cast slider with live
// recompute (auto-tapper consumes the announced XValue verbatim).
//
// S15's "mana abilities are tap-cost-only" note is half closed
// (#1215). A cost that sacrifices the SOURCE — a Treasure, a Lotus
// Petal — names no permanent and asks no question, so the planner can
// pay it; a cost that sacrifices OTHER permanents (Ashnod's Altar)
// asks which ones, and stays out of the source list. Every planned
// source still owes a {T}: the plan is a list of permanents to TAP,
// and a sacrifice-only ability (a Gold token, an Eldrazi Spawn) has no
// slot in that shape yet.
//
// A sacrifice-self source is planned LAST, behind every ordinary
// source and behind a frozen one: cracking a Treasure to pay a generic
// pip a Mountain could have paid spends a resource the player never
// agreed to spend, which is the bar the life-cost and counter-adding
// exclusions are held to.
//
// …unless the spell being cast ASKED for it. #1212's source wish
// (tapSource.Wanted) is read before the last-resort tier in both
// comparators, because a card that reads "if mana from a Treasure was
// spent to cast it" wants the Treasure cracked — that is the whole
// text. The two features meet on exactly one kind of source and the
// card's own instruction wins; the tier governs only what the planner
// reaches for when nothing asked. See the sort in
// autoTapPreferringLocked and orderUnusedByGenericPreference.

// AutoTapBudget is the maximum number of solver-recursion nodes
// the auto-tapper expands before bailing. Hit this cap and the
// public API returns (nil, false), signalling the client to fall
// back to manual tapping.
const AutoTapBudget = 10_000

// AutoTapForCost is the no-exclusions entry point — equivalent to
// AutoTapForCostExcluding(controller, cost, xValue, nil). Callers
// who haven't pre-pinned any sources should use this. The returned
// plan lists the permanent IDs in the order the solver picked them
// (colored requirements first, generic recruits second).
func (g *Game) AutoTapForCost(controller uuid.UUID, cost ParsedCost, xValue int) ([]uuid.UUID, bool) {
	return g.AutoTapForCostExcluding(controller, cost, xValue, nil)
}

// AutoTapForCostExcluding lets the caller pre-exclude a set of
// sources (lock-tap UI: the player wants Island #4 reserved for
// some other spell). Excluded sources never enter the candidate
// pool; the solver works only with what's left.
func (g *Game) AutoTapForCostExcluding(
	controller uuid.UUID,
	cost ParsedCost,
	xValue int,
	excluded map[uuid.UUID]bool,
) ([]uuid.UUID, bool) {
	var (
		plan tapPlan
		ok   bool
	)
	// A plan reads effective power and ability removal through the
	// untap-step restrictions. ReadSnapshot upgrades safely when a
	// layer pass is pending; recomputing while holding only RLock
	// would race the cache writes.
	g.ReadSnapshot(func() {
		plan, ok = g.autoTapLocked(controller, cost, xValue, excluded)
	})
	return plan.cardIDs(), ok
}

// AutoTapForCostForEffect is the *ForEffect-surface twin of
// AutoTapForCost for callers already under g.mu (the S31 legal-move
// enumerator runs inside ReadSnapshot). Read-only, same contract.
func (g *Game) AutoTapForCostForEffect(controller uuid.UUID, cost ParsedCost, xValue int) ([]uuid.UUID, bool) {
	plan, ok := g.autoTapLocked(controller, cost, xValue, nil)
	return plan.cardIDs(), ok
}

// AutoTapForCostForEffectExcluding is AutoTapForCostForEffect with
// the lock-tap exclusion set, for the one internal caller that needs
// it: an activated ability whose cost includes {T} must not fund
// itself by tapping its own source for mana first (CR 602.2b — the
// {T} and the mana are both components of the same cost). The
// legal-move enumerator has to exclude exactly what
// payAbilityManaCostLocked excludes, or it offers activations the
// engine refuses.
func (g *Game) AutoTapForCostForEffectExcluding(
	controller uuid.UUID,
	cost ParsedCost,
	xValue int,
	excluded map[uuid.UUID]bool,
) ([]uuid.UUID, bool) {
	plan, ok := g.autoTapLocked(controller, cost, xValue, excluded)
	return plan.cardIDs(), ok
}

// AutoTapForCostPreferringExcluding is AutoTapForCostExcluding with
// the spell's own SOURCE WISH (#1212): a card whose printed text
// reads which mana paid for it ("if mana from a Treasure was spent to
// cast it") would rather the plan tapped a source it can read back.
//
// The wish is an ORDERING HINT and never a filter — see
// autoTapPreferringLocked. It exists on the exported surface for one
// caller, the /autotap preview endpoint: the preview's plan is what
// the client actually taps, so a preview that ignored the wish would
// hand the player a payment the engine's own auto-tapper would not
// have made, and Hired Hexblade would draw a card through one route
// and not the other.
func (g *Game) AutoTapForCostPreferringExcluding(
	controller uuid.UUID,
	cost ParsedCost,
	xValue int,
	excluded map[uuid.UUID]bool,
	prefer ManaSourceKinds,
) ([]uuid.UUID, bool) {
	var (
		plan tapPlan
		ok   bool
	)
	g.ReadSnapshot(func() {
		plan, ok = g.autoTapPreferringLocked(controller, cost, xValue, excluded, prefer)
	})
	return plan.cardIDs(), ok
}

// plannedTap is one entry of an auto-tap plan: the permanent to tap,
// and — for a "N mana of any one color" source (#779) — the ONE
// colour the solver booked it for.
//
// The colour is carried rather than re-derived by the executor
// because one activation of such a source is one pick for all N
// tokens (ProducedManaEntry.Amounts): a Gilded Lotus the solver
// booked for {U}{U} must not mint {W} because the executor's greedy
// walk reached a different answer. That disagreement between the two
// halves of the tapper is the #273 failure, one slot wider.
//
// OneColor is empty for every ordinary source, which is all of them
// but the Lotus family.
type plannedTap struct {
	CardID   uuid.UUID
	OneColor string
}

// tapPlan is the auto-tapper's answer: the permanents to tap, in the
// order the solver picked them (coloured requirements first, generic
// recruits second). The public API projects it to a plain ID list —
// the lobby preview and the legal enumerator only ever wanted the
// IDs — and materializePlanLocked consumes the whole thing.
type tapPlan []plannedTap

// cardIDs is the projection the public auto-tap API returns.
func (p tapPlan) cardIDs() []uuid.UUID {
	if len(p) == 0 {
		return nil
	}
	out := make([]uuid.UUID, len(p))
	for i, t := range p {
		out[i] = t.CardID
	}
	return out
}

// hasCard reports whether this permanent is already in the plan.
//
// #779: one physical permanent can appear in the candidate list more
// than once — a "N mana of any one color" source contributes one
// tapSource PER OFFERED COLOUR, and they are alternatives, not
// additions. The solver would otherwise happily book a Gilded Lotus
// as three {W} AND three {U}, which is a plan the one-pick activation
// can never honour and a permanent that can only be tapped once.
func (p tapPlan) hasCard(id uuid.UUID) bool {
	for _, t := range p {
		if t.CardID == id {
			return true
		}
	}
	return false
}

// autoTapLocked is the lock-aware core. Public callers enter through
// ReadSnapshot so effective characteristics are current; internal callers
// must already hold a lock and ensure the same freshness themselves.
// Read-only — does not mutate any game state, just inspects the battlefield
// and the catalog.
func (g *Game) autoTapLocked(
	controller uuid.UUID,
	cost ParsedCost,
	xValue int,
	excluded map[uuid.UUID]bool,
) (tapPlan, bool) {
	return g.autoTapPreferringLocked(controller, cost, xValue, excluded, 0)
}

// autoTapPreferringLocked is autoTapLocked with the announcement's
// SOURCE WISH (#1212): the kinds of mana source this cast would
// rather be paid with, because its own text reads them back.
//
// THE WISH IS AN ORDERING HINT, NOT A RULE, and the bound is the
// whole design. It reaches exactly two comparators — the candidate
// sort below and orderUnusedByGenericPreference — as a tiebreak
// applied AFTER Frozen and BEFORE the existing criteria. It does not
// touch gatherTapSources's candidate set, so:
//
//   - The SOURCES are identical with and without a wish. A cost that
//     was payable stays payable and one that was not stays not, which
//     is what lets internal/legal's enumerator keep calling the
//     wishless autoTapLocked for its bool and never offer a cast the
//     gate refuses.
//   - It can never make a "spend a Treasure" card uncastable because
//     no Treasure is out. Hired Hexblade off two Swamps is an
//     ordinary 2/2 that draws no card, which is the printed
//     behaviour.
//
// The one honest caveat is the one the Frozen ordering already
// carries: solveColored shares an AutoTapBudget, so reordering can in
// principle change which plan a pathological board finds first. It
// cannot change whether one exists within an exhausted budget, since
// the search space is the same set.
//
// A zero wish is every ordinary cast and every caller that has no
// card in hand, and compares equal for everything — so the sorts are
// byte-identical to what they were before this existed.
func (g *Game) autoTapPreferringLocked(
	controller uuid.UUID,
	cost ParsedCost,
	xValue int,
	excluded map[uuid.UUID]bool,
	prefer ManaSourceKinds,
) (tapPlan, bool) {
	sources := gatherTapSources(g, controller, excluded, prefer)
	if len(sources) == 0 && (len(cost.Required) > 0 || cost.Generic+cost.XSlots*xValue > 0) {
		return nil, false
	}
	// Sort sources by restrictiveness — fewer color options first.
	// The solver picks from the front of the slice for each
	// requirement, so restrictive sources get reserved for the
	// requirements that need them most.
	sort.SliceStable(sources, func(i, j int) bool {
		// #1212: a source the spell can read back comes first. Above
		// restrictiveness rather than below it, because the point is
		// to get the wished source INTO the plan at all; the solver
		// backtracks, so preferring a less restricted source here
		// costs a little search and never an answer.
		//
		// #1215 raised it to the TOP, above the sacrifice tier below
		// and the frozen one. The wish is an explicit instruction from
		// the card being cast and the tiers are the planner's own
		// thrift, so the wish wins — and it has to, because the two
		// meet on exactly one kind of source: a Treasure is both the
		// commonest wished source and a self-sacrificing one, so a
		// wish ranked under the tier would never fire for the family
		// it was written for. See orderUnusedByGenericPreference for
		// the same order and the rest of the argument.
		if sources[i].Wanted != sources[j].Wanted {
			return sources[i].Wanted
		}
		// #1215: a source the cost EATS sorts behind everything else,
		// frozen sources included — the coloured pass takes the first
		// source that fits, so an UNWISHED Treasure is only reached
		// when no land, rock or creature on the board could have paid
		// the pip.
		if sources[i].Sacrifices != sources[j].Sacrifices {
			return !sources[i].Sacrifices
		}
		if sources[i].Frozen != sources[j].Frozen {
			return !sources[i].Frozen
		}
		return restrictivenessScore(sources[i]) < restrictivenessScore(sources[j])
	})
	need := cost.Generic + cost.XSlots*xValue
	used := make([]bool, len(sources))
	consumed := make([]int, len(sources)) // per-source slot consumption (colored reqs eat 1 each)
	plan := make(tapPlan, 0, len(cost.Required))
	budget := AutoTapBudget
	if !solveColored(sources, used, consumed, &plan, cost.Required, 0, &budget) {
		return nil, false
	}
	if budget <= 0 {
		return nil, false
	}
	if !recruitGeneric(sources, used, consumed, &plan, need) {
		return nil, false
	}
	return plan, true
}

// tapSource is one available mana ability on the battlefield. Slots
// is the parsed Produced string — a list of ProducedManaEntry,
// each with an Options set of legal colors. Sol Ring is two C
// slots; Birds of Paradise is one 5-color slot; Arcane Signet
// (Bant) is one 3-color slot (server-narrowed by commander
// identity at gather time).
type tapSource struct {
	CardID uuid.UUID
	Slots  []ProducedManaEntry
	Frozen bool

	// Sacrifices marks a source whose mana ability eats the source
	// as part of its cost (#1215) — a Treasure, a Lotus Petal. The
	// planner takes one only when nothing else can pay: it is the
	// LAST-resort tier, behind Frozen, because missing one untap is
	// a permanent coming back next turn and a cracked Treasure is a
	// resource gone for good.
	//
	// UNLESS Wanted is set. The spell that asked for Treasure mana
	// asked for the Treasure to be cracked, and its wish is read
	// first in both comparators — see the sort in
	// autoTapPreferringLocked.
	Sacrifices bool

	// OneColor is set on a candidate that exists only because its
	// permanent adds "N mana of any one color" (#779): the ONE colour
	// this candidate spends the source's single pick on, with Slots
	// already flattened to that colour's amount. A Gilded Lotus
	// contributes five such candidates, a Nyx Lotus one per colour it
	// has devotion to, and they are ALTERNATIVES — tapPlan.hasCard
	// keeps the solver from taking two of them.
	//
	// Empty on every ordinary source.
	OneColor string

	// Wanted marks a source whose kinds satisfy the announcement's
	// wish (#1212) — a Treasure, when the spell being cast reads "if
	// mana from a Treasure was spent to cast it".
	//
	// A preference and nothing more: it reorders two comparators and
	// is not consulted anywhere a source could be dropped. False on
	// every source of every ordinary cast, which is nearly all of
	// them.
	//
	// #1215: it is the FIRST key of both comparators, above
	// Sacrifices and Frozen. Until that issue the wished source could
	// not be planned at all — a Treasure was not a candidate — so the
	// hint's headline card was the one card it never reached.
	Wanted bool
}

// gatherTapSources walks the battlefield and collects every tap-
// for-mana ability the controller has access to. Tapped, foreign,
// or excluded permanents are skipped. A card with multiple mana
// abilities contributes only its first tap-cost ability for S15;
// multi-ability mana sources (Mox Diamond, City of Brass with
// activations) need a richer model that lands in a later sprint.
func gatherTapSources(g *Game, controller uuid.UUID, excluded map[uuid.UUID]bool, prefer ManaSourceKinds) []tapSource {
	if g.Battlefield == nil {
		return nil
	}
	var out []tapSource
	restrictions := g.activeUntapStepRestrictionsLocked()
	p := g.playerByIDLocked(controller)
	identity := commanderIdentityFor(g, p)
	// #1222: does anything on this board replace a mana production?
	// Asked ONCE, because the answer is no on ~every board and the
	// pricing below is a gather per colour per slot per source. See
	// producesManaReplacementsExistLocked.
	priceProduction := g.producesManaReplacementsExistLocked()
	for _, c := range g.Battlefield.Cards {
		if c.Controller != controller || c.Tapped {
			continue
		}
		if excluded[c.InstanceID] {
			continue
		}
		// S24: a permanent whose mana abilities can't be activated
		// (Arrest) is not a mana source, for the same reason a
		// gated ability isn't one below — planning it produces a
		// plan ActivateManaAbility then refuses with
		// ErrCantActivate, after the executor has already tapped
		// whatever came before it in the plan.
		if !CanActivateManaAbilities(&c) {
			continue
		}
		picked := g.autoTapAbilityFor(controller, c, ManaAbilitiesForCard(c))
		if picked == nil {
			continue
		}
		// #540: CR 302.6. A mana creature that entered this turn
		// cannot pay a {T} cost, and the auto-tapper is not a way
		// around the rule the hand-click path enforces — a Delighted
		// Halfling played this turn is not a mana source, and
		// planning it would produce a plan the executor now refuses,
		// stranding whatever it had already tapped.
		if manaTapBlockedBySickness(&c, picked) {
			continue
		}
		// #1210, CR 602.5a: the board-wide "can't be activated"
		// gate — Cursed Totem does not exempt mana abilities, so a
		// Birds of Paradise under one is not a mana source. THE
		// caller the cast gate has no equivalent of: the auto-tapper
		// never goes through ActivateManaAbility at all
		// (materializePlanLocked taps the permanent and mints its
		// mana directly), so this is not the usual "planning it would
		// produce a plan the executor refuses" — without it the plan
		// would SUCCEED and produce mana the rule forbids.
		if g.ActivationGateLocked(controller, c, ZoneBattlefield,
			ActivationAbility{Label: picked.Label, Mana: true}) != nil {
			continue
		}
		// S32 (#352): a gated ability is only a source while its
		// gate holds. Temple of the False God with four lands out
		// is not a mana source, and planning it would produce a
		// plan the executor then refuses with ErrConditionNotMet,
		// stranding whatever it had already tapped.
		if picked.Condition != nil && !picked.Condition(g, controller, c.InstanceID) {
			continue
		}
		// #789: a counter cost the planner can neither decide nor
		// afford. A Vivid land with no charge counters left is not a
		// mana source — planning it would produce a plan the
		// executor refuses, stranding whatever it had already
		// tapped, which is the same failure mode the sickness and
		// gate checks above exist to prevent.
		if !manaCounterCostPlannable(&c, picked) {
			continue
		}
		// A derived or scaled ability declares nothing useful in
		// Produced — Exotic Orchard's colours and Cabal Coffers'
		// count only exist once computed. Read-only under the lock
		// the caller already holds, same contract the activation
		// path gives the callback.
		producedStr := manaAbilityProducedLocked(g, controller, c.InstanceID, picked,
			g.maxCounterPaymentLocked(controller, c.InstanceID, picked.RemoveCounters))
		slots, err := ParseProducedMana(producedStr)
		if err != nil || len(slots) == 0 {
			continue
		}
		// Mirror the activation's option list (manaPickOptions) so
		// the planner books exactly the colours the activation will
		// offer: Birds of Paradise keeps all five with the commander's
		// identity first, and only a NarrowToCommanderIdentity source
		// (Command Tower, Arcane Signet) is intersected with the
		// identity.
		//
		// CR 903.4f (#844): that intersection can be empty — no
		// commander, or a colourless one — and a slot with nothing to
		// add is not a slot. Drop those, and a source left with no
		// slots at all is not a mana source: planning it would book a
		// mana the activation can never mint, and the executor would
		// tap the land for nothing on the way to a cast it can't pay.
		live := slots[:0]
		for _, slot := range slots {
			slot.Options = manaPickOptions(slot.Options, identity, picked.NarrowToCommanderIdentity)
			if len(slot.Options) == 0 {
				continue
			}
			live = append(live, slot)
		}
		slots = live
		if len(slots) == 0 {
			continue
		}
		// #1222: what this source will REALLY make. "If you tap a
		// permanent for mana, it produces twice as much of that mana
		// instead" (Mana Reflection) is a CR 614 replacement, so a
		// planner that booked the printed amount would tap two lands
		// where a Mana-Reflected one pays — and, worse, would read a
		// payable cast as unpayable in strict mode. Priced through the
		// same gathered AppliesTo / Replace pairs the executor runs, so
		// the two halves of the tapper cannot disagree about what a
		// source produces, which is the rule this whole function is
		// built around.
		//
		// BEFORE the #1212 wish below, because this is the branch that
		// can still drop the source: a permanent the window leaves with
		// nothing to add is not a mana source, wished-for or not.
		if priceProduction {
			slots = g.priceProducedSlotsLocked(controller, c.InstanceID, slots)
			if len(slots) == 0 {
				// The window replaces this source's production away
				// entirely. Tapping a land for no mana is worse than
				// not tapping it — the CR 903.4f posture, one line up.
				continue
			}
		}
		// #1212: the wish, answered off the same snapshot the mana
		// this permanent produces will carry (manaSourceKindsOf), so
		// the planner and the record can never disagree about what
		// this source is. The AMOUNT it produces is priced above and
		// the KINDS it produces are unaffected by that — CR 106.12b
		// replaces how much mana is produced, not what produced it.
		wanted := prefer != 0 && manaSourceKindsOf(c).HasAny(prefer)
		out = appendTapSource(out, c.InstanceID, slots,
			untapStepRestrictedBy(&c, g, restrictions) || c.hasNextUntapSkipFor(controller),
			wanted, picked.SacrificeCost)
	}
	return out
}

// Return values of oneColorSlot.
const (
	// oneColorSlotNone: no slot is a "N mana of any one color" pick,
	// which is every source but the Lotus family.
	oneColorSlotNone = -1
	// oneColorSlotUnplannable: more than one such slot on one
	// ability. Two independent one-colour picks would be a cross
	// product of candidates for a shape no printed card has, so the
	// planner declines the source entirely and the player taps it by
	// hand — the same weaker-and-safe direction the restricted-output
	// exclusion takes.
	oneColorSlotUnplannable = -2
)

// oneColorSlot returns the index of the source's single "N mana of
// any one color" slot (ProducedManaEntry.OneColorAmounts), or one of
// the constants above.
func oneColorSlot(slots []ProducedManaEntry) int {
	found := oneColorSlotNone
	for i, slot := range slots {
		if !slot.OneColorAmounts() {
			continue
		}
		if found != oneColorSlotNone {
			return oneColorSlotUnplannable
		}
		found = i
	}
	return found
}

// appendTapSource turns one permanent's parsed slots into the
// candidate sources the solver may pick from.
//
// Usually that is exactly one candidate. #779: a permanent that adds
// "N mana of any one color" (Gilded Lotus, Lotus Field, Nyx Lotus,
// White Lotus Tile) contributes ONE CANDIDATE PER OFFERED COLOUR,
// each with the pick already flattened into that colour's amount —
// three {U} slots for a Gilded Lotus booked blue, four {G} slots for
// a Nyx Lotus with devotion G4. The solver then reasons about them
// with the model it already has (one slot, one mana, one colour), and
// the candidates are mutually exclusive because they name the same
// CardID (tapPlan.hasCard).
//
// This replaces #742's blanket skip. The skip was the weaker
// direction — a payable cast read as unpayable in strict mode, in the
// cast preview and to every bot (#779) — and the expansion is the
// stronger one that the one-pick activation can still honour, because
// each candidate spends the pick on exactly one colour.
//
// Surplus is not a problem the planner has to solve: a Gilded Lotus
// booked for {U}{U} leaves its third slot in the spare tally, which
// recruitGeneric spends on the generic half of the same cost, and
// anything still left floats (CR 106.4).
func appendTapSource(out []tapSource, cardID uuid.UUID, slots []ProducedManaEntry, frozen, wanted, sacrifices bool) []tapSource {
	idx := oneColorSlot(slots)
	switch idx {
	case oneColorSlotNone:
		return append(out, tapSource{
			CardID:     cardID,
			Slots:      slots,
			Frozen:     frozen,
			Wanted:     wanted,
			Sacrifices: sacrifices,
		})
	case oneColorSlotUnplannable:
		return out
	}
	pick := slots[idx]
	for _, color := range pick.Options {
		n := pick.AmountFor(color)
		if n <= 0 {
			// A zero-amount option adds nothing; ParseProducedMana
			// already drops those, so this is belt-and-braces.
			continue
		}
		variant := make([]ProducedManaEntry, 0, len(slots)-1+n)
		variant = append(variant, slots[:idx]...)
		for k := 0; k < n; k++ {
			variant = append(variant, ProducedManaEntry{Options: []string{color}})
		}
		variant = append(variant, slots[idx+1:]...)
		out = append(out, tapSource{
			CardID:     cardID,
			Slots:      variant,
			Frozen:     frozen,
			Wanted:     wanted,
			Sacrifices: sacrifices,
			OneColor:   color,
		})
	}
	return out
}

// autoTapAbilityFor picks the one mana ability the auto-tapper is
// willing to fire on a permanent, or nil when none qualifies. Shared
// by the planner (gatherTapSources) and the executor
// (materializePlanLocked) so the two can never disagree about which
// ability index a planned card is going to be tapped for.
//
// A METHOD since #1183, because one of the exclusions is a fact about
// the game rather than about the ability shape: an exhaust ability
// this OBJECT has already activated is not a mana source. Putting it
// here rather than beside the sickness and gate checks at the two call
// sites is the point — a card with two mana abilities whose FIRST is a
// spent exhaust still auto-taps the second, which a per-source check
// outside the picker would have got wrong.
//
// Seven exclusions, all for the same reason — the auto-tapper's
// contract is "no further player decisions and no hidden costs":
//
//   - a SPENT EXHAUST ability cannot be activated at all (#1183), and
//     the planner must not book mana the executor would then refuse to
//     mint. It is the strongest of the exclusions: the others are the
//     planner declining a decision it may not make, this one is the
//     rule;
//
//   - a sacrifice cost that names OTHER permanents (Ashnod's Altar's
//     "Sacrifice a creature") needs a permanent named, and naming one
//     is a decision the planner may not make. #1215: this used to read
//     `a.SacrificeCost`, which is the OTHER clause — the one that eats
//     the SOURCE (a Treasure, a Lotus Petal) and needs no decision at
//     all, because ActivateManaAbility pays it off the source's own
//     ID. The two were folded into one exclusion and the half that
//     asks nothing was excluded with the half that asks. A board of
//     Treasures read as unpayable to the cast preview, to the strict
//     gate and to every bot, and #1212's source wish could not reach
//     the one family it was built for;
//
//   - a cost that ADDS a counter spends a resource the player never
//     agreed to spend, like a life cost (#789);
//
//   - a life cost spends a resource the player never agreed to spend
//     (Mana Confluence);
//
//   - a rider spends one too, one the player can't decline (Ancient
//     Tomb's 2 damage);
//
//   - a MANA cost is recursive (the Signet cycle, Cabal Coffers): the
//     planner would have to solve a second cost to fund the first,
//     and the activation path deliberately refuses to auto-tap into a
//     mana ability anyway. Signets stay hand-activated — unless a
//     CR 601.2f modifier prices the component away to nothing (#1191:
//     Boom Scholar's exhaust discount reaches Loot, the Pathfinder's
//     "{G}, {T}"), in which case there is no second cost left to
//     solve and the source is plannable exactly as if it had never
//     printed one;
//
//   - RESTRICTED output is a decision, not a resource (Ancient
//     Ziggurat, Eldrazi Temple, Delighted Halfling's coloured half).
//     Spending a restricted token on the cast in front of you may be
//     right or may waste the only mana that could have cast the
//     creature you were saving it for, and the planner cannot know
//     which. It plans around them; the player clicks them.
//
// The last two are simplifications, both in the WEAKER-than-printed
// direction: the cards are fully activatable by hand from the
// permanent's ability menu, and mana already floated from them is
// spent by the cast path like any other (that is the point of the
// spend context — see mana_restriction.go).
//
// The painland and Talisman cycles come through this filter intact,
// because on those cards the painless "{T}: Add {C}" is ability 0 and
// the rider lives on ability 1 — the auto-tapper plans them as
// colorless sources and leaves the painful colored half to a
// deliberate click. Ancient Tomb and Mana Confluence have no painless
// ability and drop out of auto-tap planning entirely; they are still
// fully activatable by hand from the permanent's ability menu.
//
// City of Brass deliberately does NOT drop out: its pain is a
// separate "whenever this land becomes tapped" TRIGGER, not part of
// the mana ability, and it fires on the EventTapCard that
// materializePlanLocked emits just as it does on a hand-clicked
// activation. Auto-tapping it costs the printed point of damage,
// which is the right answer.
//
// S15's other standing limitation is unchanged: one ability per card,
// because a tapSource that offered two would let the solver tap the
// same permanent twice.
//
// `source` is the permanent itself, not just its ID (#1191): pricing
// a ManaCost component through the CR 601.2f pass needs a Card to
// build the CostQuery from, exactly as abilityCostQueryLocked does for
// a CR 602 ability.
//
// Caller must hold g.mu.
func (g *Game) autoTapAbilityFor(asker uuid.UUID, source Card, abilities []ManaAbilityShape) *ManaAbilityShape {
	for i := range abilities {
		a := abilities[i]
		// Every planned source owes a {T}: a tapPlan is a list of
		// permanents to tap, so a sacrifice-ONLY ability (a Gold
		// token, an Eldrazi Spawn) has no slot in it yet — see the
		// file header.
		if !a.TapCost {
			continue
		}
		// #1215: the sacrifice-OTHER half only. Sacrificing the source
		// itself is plannable — see the bullet above.
		if a.SacrificeOther != nil {
			continue
		}
		// #1183: "Activate each exhaust ability only once", and this
		// object has. Not a decision the planner is declining — the
		// activation path would refuse it with ErrAbilityExhausted,
		// so planning it would strand whatever the plan had already
		// tapped, which is the failure mode every check around this
		// one exists to prevent.
		if g.ManaAbilityExhausted(asker, source.InstanceID, a) {
			continue
		}
		if a.LifeCost > 0 || a.Rider != nil {
			continue
		}
		if len(a.Restrictions) > 0 || a.RestrictionsFunc != nil {
			continue
		}
		if a.ManaCost != "" {
			// #1191: priced rather than read off the printed string —
			// see the file header's MANA-cost bullet. Spending the
			// controller's floated mana on THIS activation is still a
			// decision the planner cannot make, so anything left to
			// pay after the CR 601.2f pass still drops the source; an
			// unparseable cost or a modifier error is the same
			// refusal a printed cost would have given the activation
			// path, so it is treated as "still owes something" rather
			// than plannable.
			priced, err := g.ManaAbilityManaCostForEffect(asker, source, a)
			if err != nil || !priced.Empty() {
				continue
			}
		}
		// #789: a cost that PUTS a counter on the source spends a
		// resource the player never agreed to spend, exactly as a
		// life cost does — Devoted Druid's -1/-1 is permanent and
		// the planner must not decide to take it. A cost that
		// REMOVES counters may be plannable; whether this particular
		// one is depends on the permanent, so it is asked separately
		// by manaCounterCostPlannable.
		if a.AddCounter != nil {
			continue
		}
		// #758: a cost that taps OTHER permanents spends a resource
		// the player was never asked about — auto-tapping Springleaf
		// Drum would tap a creature that was being kept back to
		// block. The same bar the life cost fails, and the planner
		// has no way to weigh it.
		if !a.TapOthers.Empty() {
			continue
		}
		// #1213: a discard cost, on the same ground one zone over —
		// auto-tapping Skirge Familiar would pitch a card the player
		// never offered, and WHICH card is a decision the planner
		// makes none of. A hand-clicked Skirge Familiar is a mana
		// source; an auto-tapped one is not.
		if a.DiscardCards != nil {
			continue
		}
		return &a
	}
	return nil
}

// manaCounterCostPlannable reports whether the auto-tapper may plan
// this source for this ability given its RemoveCounters component
// (#789). The one copy of the rule the planner (gatherTapSources)
// and the executor (materializePlanLocked) share, for the same
// reason autoTapAbilityFor and manaTapBlockedBySickness are shared:
// two copies drift, and when the planner's is the laxer one the
// executor strands whatever the plan had already tapped.
//
// Three conditions, and all three are the auto-tapper's standing
// contract — "no further player decisions, no hidden costs" —
// applied to counters:
//
//   - The counters must come off the SOURCE. "Remove a counter from
//     a creature you control" and "from among artifacts you control"
//     both ask which permanent pays, and the planner answers no
//     questions.
//   - The kind and the count must be PRINTED. An any-kind cost asks
//     which kind; a variable cost ("Remove X storage counters", "any
//     number") asks how many, and the answer changes how much mana
//     arrives — the player decides that, not the planner.
//   - The permanent must hold enough RIGHT NOW. This is the "never
//     plan a Vivid land with no charge counters" rule, and it is
//     re-asked by the executor because a plan can arrive stale.
//
// Vivid Creek and Vivid Grove pass all three, which is the point:
// the commonest counter-cost mana ability in the game auto-taps like
// any other land until its second charge counter is gone, and then
// quietly stops being a five-colour source — exactly as it stops
// being one in paper.
func manaCounterCostPlannable(c *Card, ab *ManaAbilityShape) bool {
	if ab == nil || c == nil {
		return false
	}
	if ab.AddCounter != nil {
		return false
	}
	rc := ab.RemoveCounters
	if rc == nil {
		return true
	}
	if rc.From != nil || rc.Among || rc.Variable || rc.Counter == "" || rc.N < 1 {
		return false
	}
	return c.Counters[rc.Counter] >= rc.N
}

// manaTapBlockedBySickness reports whether CR 302.6 forbids tapping
// this permanent for the given mana ability right now. The one copy
// of the rule the auto-tapper's two halves share — the planner
// (gatherTapSources) and the executor (materializePlanLocked) — for
// the same reason autoTapAbilityFor is shared: two copies of a
// legality rule drift, and when the planner's copy is the laxer one
// the executor strands whatever the plan had already tapped.
//
// Only a {T} cost is gated: an ability with no tap symbol (a
// sacrifice or mana cost) is unaffected by summoning sickness, and
// CR 302.6 is about CREATURES — a Sol Ring, a Treasure and a
// fetchland all tap the turn they arrive. Haste is the read-time
// bypass, handled inside HasSummoningSickness. The IsCreature prefix
// is redundant since #530 and stays as documentation, matching
// ActivateManaAbility's copy of the same gate.
//
// Layers must be fresh: both creature-hood and haste are effective
// characteristics. The planner enters through ReadSnapshot and the
// executor's callers already owe the same freshness for
// ManaAbilitiesForCard.
func manaTapBlockedBySickness(c *Card, ab *ManaAbilityShape) bool {
	if c == nil || ab == nil || !ab.TapCost {
		return false
	}
	return c.IsCreature() && HasSummoningSickness(c)
}

// restrictivenessScore lower = more restrictive (better picked
// first). For multi-slot sources, the score is the sum of per-slot
// option counts — Sol Ring [{C},{C}] = 2, Forest [{G}] = 1,
// Birds [{W,U,B,R,G}] = 5, Arcane Signet (Bant) [{W,U,G}] = 3.
func restrictivenessScore(s tapSource) int {
	n := 0
	for _, slot := range s.Slots {
		n += len(slot.Options)
	}
	return n
}

// solveColored places each colored requirement onto a source slot
// via depth-first backtracking. Returns true on a complete
// assignment, false on dead-end. The shared budget counter halts
// pathological searches.
func solveColored(
	sources []tapSource,
	used []bool,
	consumed []int,
	plan *tapPlan,
	reqs []ColorRequirement,
	reqIdx int,
	budget *int,
) bool {
	if *budget <= 0 {
		return false
	}
	*budget--
	if reqIdx >= len(reqs) {
		return true
	}
	req := reqs[reqIdx]
	for i := range sources {
		// Try each source — used or not. A used source can still
		// satisfy a requirement from one of its remaining slots.
		if consumed[i] >= len(sources[i].Slots) {
			continue
		}
		// #779: a colour candidate of a permanent already in the plan
		// under a DIFFERENT colour is not available — the permanent
		// taps once and makes one pick.
		if !used[i] && plan.hasCard(sources[i].CardID) {
			continue
		}
		// Find a slot in this source that hasn't been consumed
		// AND matches the requirement's options.
		slotIdx := pickMatchingSlot(sources[i], consumed[i], req.Options)
		if slotIdx < 0 {
			continue
		}
		// Tentatively pick. Mark used (idempotent for
		// already-used sources), bump the consumed counter,
		// append to plan only on the source's first use.
		wasUsed := used[i]
		if !wasUsed {
			used[i] = true
			*plan = append(*plan, plannedTap{CardID: sources[i].CardID, OneColor: sources[i].OneColor})
		}
		consumed[i]++
		if solveColored(sources, used, consumed, plan, reqs, reqIdx+1, budget) {
			return true
		}
		consumed[i]--
		if !wasUsed {
			used[i] = false
			*plan = (*plan)[:len(*plan)-1]
		}
	}
	return false
}

// pickMatchingSlot returns the index of the first slot in `s`
// (starting at startIdx — the next un-consumed slot) whose Options
// intersect with reqOptions. -1 when nothing matches. The
// "starting at consumed[i]" convention is fine for the simple
// uniform-slot case (Sol Ring's two C slots are interchangeable);
// a richer multi-color mana rock would need a per-slot pick, but
// that doesn't ship in S15.
func pickMatchingSlot(s tapSource, startIdx int, reqOptions []string) int {
	for i := startIdx; i < len(s.Slots); i++ {
		slot := s.Slots[i]
		for _, opt := range slot.Options {
			if matchColor(opt, reqOptions) {
				return i
			}
		}
	}
	return -1
}

// recruitGeneric tallies spare slots across already-used sources
// and recruits additional unused sources until total spare meets
// `need`. Generic-friendly preference: colorless-producing slots
// first (preserves colored mana for the next cast), then
// any-color slots, then plain colored. Returns true on success,
// false when even the full source pool can't cover the need.
func recruitGeneric(
	sources []tapSource,
	used []bool,
	consumed []int,
	plan *tapPlan,
	need int,
) bool {
	if need <= 0 {
		return true
	}
	// Spare slots from already-used sources first — they're
	// already paid for.
	spare := 0
	for i := range sources {
		if !used[i] {
			continue
		}
		spare += len(sources[i].Slots) - consumed[i]
	}
	if spare >= need {
		return true
	}
	deficit := need - spare
	// Recruit additional sources, preferring colorless producers.
	candidateOrder := orderUnusedByGenericPreference(sources, used)
	for _, i := range candidateOrder {
		if used[i] {
			continue
		}
		// #779: same exclusion the coloured pass makes — a colour
		// candidate whose permanent is already tapped by this plan is
		// not a second source.
		if plan.hasCard(sources[i].CardID) {
			continue
		}
		used[i] = true
		*plan = append(*plan, plannedTap{CardID: sources[i].CardID, OneColor: sources[i].OneColor})
		deficit -= len(sources[i].Slots)
		if deficit <= 0 {
			return true
		}
	}
	if deficit > 0 {
		return false
	}
	return true
}

// orderUnusedByGenericPreference returns indices into `sources`
// for the unused entries in the order the generic recruiter
// should try them. Colorless-only sources (Sol Ring) come first
// — they preserve colored mana for future casts. Then any-color
// (Birds-style), then plain monocolored. Within a tier, slot
// count descending so a single recruitment plan covers more
// generic faster. A frozen source comes after every ordinary source
// regardless of its generic tier, preserving a permanent that will
// miss its next normal untap unless no other source can pay.
//
// #1215: and a source the cost EATS comes after even a frozen one.
// The generic half of a cost is exactly where the wrong answer is
// cheapest to reach — a Treasure is a five-colour source, so the
// any-colour tier would otherwise recruit it ahead of the basic land
// sitting untapped beside it — and it is also where it costs the
// most, because the Treasure is gone and the land untaps.
func orderUnusedByGenericPreference(sources []tapSource, used []bool) []int {
	type rank struct {
		idx        int
		wanted     bool
		sacrifices bool
		frozen     bool
		tier       int // 0 = colorless-only, 1 = any-color, 2 = monocolored
		slotCnt    int
	}
	out := make([]rank, 0, len(sources))
	for i, s := range sources {
		if used[i] {
			continue
		}
		t := tierForGeneric(s)
		out = append(out, rank{
			idx:        i,
			wanted:     s.Wanted,
			sacrifices: s.Sacrifices,
			frozen:     s.Frozen,
			tier:       t,
			slotCnt:    len(s.Slots),
		})
	}
	sort.SliceStable(out, func(a, b int) bool {
		// #1212: this is where the wish actually bites. A Treasure is
		// an any-colour source (tier 1) and a Sol Ring is colourless
		// (tier 0), so without the hint a Hired Hexblade's generic
		// pip is always paid by the Sol Ring and the card never
		// draws. Above the tier, because the tier is exactly the
		// preference being overridden.
		//
		// #1215 moved it to the TOP, above the sacrifice tier and the
		// frozen one. The wish and the tier meet on exactly one kind
		// of source — a Treasure is both the commonest wished source
		// and a self-sacrificing one — so a wish ranked below the
		// tier is a wish that never fires for the family it was
		// written for. And a wish that is allowed to beat "this
		// permanent will be DESTROYED" must beat "this permanent
		// misses one untap" too, which is strictly the cheaper of the
		// two. One sentence for the whole order: the card being cast
		// gets the source it asks for, and where it asks for nothing
		// the planner spends the cheapest thing on the board.
		if out[a].wanted != out[b].wanted {
			return out[a].wanted
		}
		if out[a].sacrifices != out[b].sacrifices {
			return !out[a].sacrifices
		}
		if out[a].frozen != out[b].frozen {
			return !out[a].frozen
		}
		if out[a].tier != out[b].tier {
			return out[a].tier < out[b].tier
		}
		return out[a].slotCnt > out[b].slotCnt
	})
	idxs := make([]int, len(out))
	for i, r := range out {
		idxs[i] = r.idx
	}
	return idxs
}

// tierForGeneric scores a source's "generic-spending preference"
// — lower = recruit-first. Colorless-only producers are tier 0
// (Sol Ring); any-color (Birds, Signet) tier 1; plain mono-
// colored tier 2 (basics — preserve them as long as possible).
func tierForGeneric(s tapSource) int {
	allColorless := true
	anyMulti := false
	for _, slot := range s.Slots {
		if len(slot.Options) == 1 && slot.Options[0] == "C" {
			continue
		}
		allColorless = false
		if len(slot.Options) > 1 {
			anyMulti = true
		}
	}
	switch {
	case allColorless:
		return 0
	case anyMulti:
		return 1
	default:
		return 2
	}
}

// intersectColors returns the elements of `a` that also appear in
// `b`, in `a`'s order. Used by manaPickOptions: Arcane Signet's
// commander-identity narrowing (the produced "{W|U|B|R|G}"
// intersected with {W,U,G} for a Bant deck) and the identity-first
// ordering every other multi-option slot gets.
func intersectColors(a, b []string) []string {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	bSet := make(map[string]struct{}, len(b))
	for _, x := range b {
		bSet[x] = struct{}{}
	}
	out := make([]string, 0, len(a))
	for _, x := range a {
		if _, ok := bSet[x]; ok {
			out = append(out, x)
		}
	}
	return out
}

// hasMultiOptionSlot reports whether any slot is a pipe — a colour the
// controller picks. Used to decide whether reading the commander's
// colour identity can change anything at all: the read walks every
// zone, and a Forest's "{G}" has no use for it.
func hasMultiOptionSlot(slots []ProducedManaEntry) bool {
	for _, slot := range slots {
		if len(slot.Options) > 1 {
			return true
		}
	}
	return false
}

// colorOffered reports whether `color` is one of the options a slot
// offers. The executor's staleness check for a planned one-colour
// pick (#779): a plan built before a Nyx Lotus's devotion moved may
// name a colour the source no longer offers.
func colorOffered(options []string, color string) bool {
	if color == "" {
		return false
	}
	for _, o := range options {
		if o == color {
			return true
		}
	}
	return false
}
