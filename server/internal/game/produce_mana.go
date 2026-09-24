package game

import (
	"errors"

	"github.com/google/uuid"
)

// produce_mana.go is the CR 614 window on the MANA PRODUCED (#1222).
//
// # What was missing, precisely
//
// Not the mana pipeline. Since #763 every mana in the game arrives
// through one of four sites and each of them already knows the colour
// it is adding, the restrictions it rides with and the player it
// belongs to. What there was no event for is the AMOUNT — "if a
// permanent you control would produce one or more mana, it produces
// twice as much of that mana instead" (Mana Reflection) is not a fact
// about a token going into a pool, it is a replacement of the
// production, and CR 106.12b is its own clause for exactly that
// reason.
//
// So RepEventProduceMana is opened once per production, before any of
// it is in the pool, the way RepEventMill is opened once per mill
// instruction and RepEventCreateTokens once per creation
// (ADR 0061 decision 1, ADR 0013 §5u). It carries COLOURS rather than
// a bare count because the cards say "twice as much of THAT mana".
//
// # The four sites, and why the colour is always settled
//
//	ActivateManaAbility      a single-option slot ("{T}: Add {G}")
//	ResolveManaChoice        a pipe slot's answered pick (Birds of
//	                         Paradise, Gilded Lotus's three of one)
//	materializePlanLocked    the auto-tap executor's own taps
//	addManaSlotsLocked       a spell's "Add {B}{B}{B}" (Dark Ritual)
//	                         and #763's triggered mana abilities
//
// All four reach produceManaLocked, which is the only place in the
// engine that mints a ManaToken and emits EventManaAdded. A slot whose
// colour is NOT yet known — a pipe waiting on a pick — does not open a
// window at its activation; the pick does, when it is answered. That
// is the same reason ADR 0074 §3 fires the triggered mana abilities
// from ResolveManaChoice: it is the only moment a Birds-style source's
// colour exists.
//
// # It can never pause
//
// Every event of this kind sets mustSettleNow, and this is the one
// kind for which that is a rules fact rather than a cost argument.
// CR 605.3b makes activating a mana ability a single indivisible step
// with no stack and no priority window inside it, so there is no point
// between paying the cost and producing the mana at which a player
// could be asked anything; and the auto-tapper's contract is "no
// further player decisions required", so a prompt raised halfway
// through materializePlanLocked would strand a half-tapped board in
// the middle of a cast.
//
// What that costs is the CR 616.1 ordering choice when two production
// replacements share a window. It is unobservable for both printed
// cards — Mana Reflection doubles and Nyxbloom Ancient triples, and
// x2 then x3 is x6 in either order — and the gathered order still
// composes through the apply-loop, one effect per pass, so the totals
// are right. See ADR 0013 §5ab.
//
// # And what that buys
//
// No tail. A mill, a token creation and a keyword action all carry one
// because their window can pause and the resume has to finish what the
// caller asked for; a production that cannot pause owes nothing to
// anybody across a boundary that does not exist. The three switches
// that name what a kind owes on a pause, a cancellation or a dropped
// prompt therefore all say "nothing, and here is why".
//
// # The planner prices through the same predicate
//
// A Mana Reflection board pays more per land, so the auto-tapper has to
// PLAN with the replaced amount or it taps twice what it needs (and,
// worse, reads a payable cast as unpayable in strict mode). The
// planner cannot open a real window — it runs under a read lock and
// must not touch the once-per-event map — so it prices with
// producedManaPreviewLocked, which walks exactly the same gathered
// AppliesTo / Replace pairs against a scratch event and keeps its own
// applied set. One predicate, two readers, no second copy of the rule.

// produceManaLocked is the ONE body that puts mana in a pool. It opens
// the CR 614 window on the production, mints one ManaToken per mana the
// window settles on and emits one EventManaAdded per token, and
// returns the colours that actually arrived — which is what the four
// call sites feed to #763's triggered mana abilities.
//
// `colors` is the production as the ability or effect declared it, one
// entry per mana, every entry a settled colour. `restrictions` is the
// spend restriction each minted token carries; it is copied per token,
// because the catalog's slice is process-lifetime and a ManaToken is
// game state clone.go deep-copies for undo (#259).
//
// `kinds` is #1212's snapshot of WHAT the producing permanent was when
// it produced (`ManaToken.SourceKinds`), taken by the caller because
// only the caller knows which object to snapshot: the permanent as it
// was when it was tapped for the executor, the one the choice recorded
// for an answered pick, a live lookup for a spell. Collapsing the four
// mint sites into this body moved the STAMPING here and left the
// SNAPSHOTTING where #1212 put it, which is the half that has to stay
// per-site — and the replaced mana carries the same kinds as the
// printed mana, because CR 106.12b replaces how much mana is produced,
// not what produced it.
//
// `fromTap` says this production is part of tapping a permanent for
// mana (CR 106.12a) — the printed condition on Mana Reflection and
// Nyxbloom Ancient. See ReplacementEvent.ManaFromTap.
//
// `pending` is the auto-tapper's remaining colour requirements, or nil
// outside a cast. When it is set, this books one requirement for each
// mana the WINDOW ADDED beyond what the caller asked for: every one of
// those callers has already booked one requirement per mana the ability
// PRINTS (that is what pickColorForSlot and bookColorRequirement do at
// the call site), and the replaced surplus has to be booked too or the
// next slot re-pays a pip these tokens already cover (#273).
//
// Caller must hold g.mu in write mode.
func (g *Game) produceManaLocked(
	p *Player,
	source uuid.UUID,
	colors []string,
	restrictions []string,
	kinds ManaSourceKinds,
	fromTap bool,
	pending *[]ColorRequirement,
) []string {
	if p == nil || len(colors) == 0 {
		return nil
	}
	produced := g.replaceProducedManaLocked(p.ID, source, colors, fromTap)
	for _, color := range produced {
		if color == "" {
			continue
		}
		p.ManaPool.AddMana(ManaToken{
			Color:        color,
			Source:       source,
			Restrictions: copyRestrictions(restrictions),
			SourceKinds:  kinds,
		})
		g.EmitEvent(Event{Kind: EventManaAdded, Actor: p.ID, Source: source, Colors: []string{color}})
	}
	if pending != nil {
		// Only the surplus: see the doc comment. Taken off the tail
		// because a multiplication repeats each entry in place, so the
		// tail is exactly the mana the caller did not book.
		for i := len(colors); i < len(produced); i++ {
			bookColorRequirement(produced[i], pending)
		}
	}
	return produced
}

// replaceProducedManaLocked runs the CR 614 window on one production
// and returns the mana it settles on, or nil for a production replaced
// away entirely (CR 614.10).
//
// Caller must hold g.mu in write mode.
func (g *Game) replaceProducedManaLocked(playerID, source uuid.UUID, colors []string, fromTap bool) []string {
	ev := &ReplacementEvent{
		Kind:        RepEventProduceMana,
		Actor:       playerID,
		Source:      source,
		ManaPlayer:  playerID,
		ManaSource:  source,
		ManaFromTap: fromTap,
		// Its own backing array. The caller's slice is often the
		// catalog's or a loop variable's, and a doubler rewrites this
		// field in place.
		ManaColors: append([]string(nil), colors...),
		// CR 605.3b: no prompt can be raised inside a mana ability's
		// resolution, and the auto-tapper may not raise one at all.
		// See the file comment.
		mustSettleNow: true,
	}
	out, err := g.applyReplacementsLocked(ev)
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		// mustSettleNow forecloses errReplacementPending, so the only
		// way here is a genuinely broken pipeline. The production goes
		// through unreplaced — weaker than printed, never stronger, and
		// never a pool that silently lost its mana.
		g.clearReplacementEventLocked(ev.ID)
		return colors
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	return out.ManaColors
}

// producedManaPreviewLocked is the READ-ONLY half of the same window:
// how much mana a production of `colors` by `source` would actually
// become, asked without touching the game.
//
// The auto-tap PLANNER is its only caller, and it needs exactly this:
// gatherTapSources runs inside ReadSnapshot under a read lock, so it
// may neither mint an event ID nor write the once-per-event map that
// applyReplacementsLocked keeps — and a planner that ignored the window
// would tap two lands for a cost one Mana-Reflected land pays.
//
// It walks the SAME gathered effects the real window walks
// (gatherActiveReplacementsLocked) and fires the SAME Replace
// functions, with its own applied set standing in for the map, so the
// planner and the executor cannot drift apart. That is sound only
// because a RepEventProduceMana replacement rewrites the event and
// nothing else — the contract written on ReplacementEvent.ManaColors,
// and the reason the preview skips any effect that would ask its
// controller a question (there is nobody to ask during planning, and
// the real window skips them too, under mustSettleNow).
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) producedManaPreviewLocked(playerID, source uuid.UUID, colors []string, fromTap bool) []string {
	if len(colors) == 0 {
		return nil
	}
	ev := &ReplacementEvent{
		Kind:       RepEventProduceMana,
		Actor:      playerID,
		Source:     source,
		ManaPlayer: playerID,
		ManaSource: source,
		// CR 106.12a: whether this production is a TAP for mana. True
		// for every battlefield source the planner gathers, and false
		// for the one kind that is not a permanent at all — a Spirit
		// Guide exiled out of a hand (#1228). It decides whether
		// "whenever you tap a permanent for mana" replacements see
		// the production at all, so passing the wrong one would price
		// a Mana-Reflected Spirit Guide at two {R}.
		ManaFromTap:   fromTap,
		ManaColors:    append([]string(nil), colors...),
		mustSettleNow: true,
	}
	// ev.ID stays zero, so gatherActiveReplacementsLocked reads a nil
	// applied map and re-offers everything each pass; this is the set
	// that stands in for it (CR 614.5).
	applied := make(map[ReplacementEffectID]bool, 2)
	for iter := 0; iter < maxReplacementIters; iter++ {
		if ev.Canceled {
			return nil
		}
		chosen := -1
		gathered := g.gatherActiveReplacementsLocked(ev)
		for i := range gathered {
			if applied[gathered[i].id] || asksItsOwnQuestion(gathered[i].effect) {
				continue
			}
			chosen = i
			break
		}
		if chosen < 0 {
			break
		}
		// CR 616.1f: one effect, then re-gather — the same loop shape
		// applyFirstGatheredLocked gives the un-prompted paths, and the
		// same order (the gathered one) mustSettleNow settles on.
		applied[gathered[chosen].id] = true
		if gathered[chosen].effect.Replace != nil {
			// The error is dropped rather than logged: emitting would
			// be a write, and this is the read-only path. A misbehaving
			// replacement is reported by the real window when the mana
			// is actually produced.
			_ = gathered[chosen].effect.Replace(ev, g, gathered[chosen].source)
		}
	}
	if ev.Canceled {
		return nil
	}
	return ev.ManaColors
}

// producesManaReplacementsExistLocked reports whether ANY replacement
// on the board watches a mana production. It is the planner's cheap
// gate: pricing every slot of every source costs a gather per colour
// per slot, and on the ~every board where no such card is out the
// answer is no and the planner skips the whole pass.
//
// Deliberately asks only the WATCH key and not AppliesTo, which is the
// same pre-filter gatherActiveReplacementsLocked applies first: a false
// here means no effect of the kind can possibly apply, and a true means
// the full pricing pass is worth running.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) producesManaReplacementsExistLocked() bool {
	for i := range g.BuiltinReplacements {
		if eventKindMatches(g.BuiltinReplacements[i].Watches, RepEventProduceMana) {
			return true
		}
	}
	for i := range g.TurnScopedReplacements {
		if eventKindMatches(g.TurnScopedReplacements[i].Watches, RepEventProduceMana) {
			return true
		}
	}
	for i := range g.testReplacements {
		if eventKindMatches(g.testReplacements[i].Watches, RepEventProduceMana) {
			return true
		}
	}
	if CatalogReplacements == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		for _, eff := range CatalogReplacements(CatalogAbilityKey(g.Battlefield.Cards[i])) {
			if eventKindMatches(eff.Watches, RepEventProduceMana) {
				return true
			}
		}
	}
	return false
}

// priceProducedSlotsLocked rewrites a source's produced-mana slots with
// the amounts the CR 614 window would settle on, so the auto-tap
// planner books what the source will REALLY make.
//
// The rewriting is expressed in the grammar the planner already
// understands (ProducedManaEntry), which is what keeps the change to
// one function:
//
//   - a single-option slot that doubles becomes TWO slots of that
//     colour, exactly as ParseProducedMana already expands "{G2}";
//   - a multi-option slot that doubles becomes ONE slot with per-option
//     Amounts — a "two mana of any one color" pick, which is #779's
//     shape and which appendTapSource already expands into one
//     candidate per colour. That is the honest model: a Mana-Reflected
//     Birds of Paradise is still ONE pick, so it cannot pay {W}{U}.
//   - a slot the window empties is dropped, and a source left with no
//     slots is not a mana source (gatherTapSources drops it before
//     tapping it, the CR 903.4f posture).
//
// Slots are returned unchanged when nothing changes, which is every
// board with no production replacement on it.
//
// The EXECUTOR does not price. It produces from the printed slots and
// lets the real window apply the amount (produceManaLocked) — pricing
// there as well would multiply twice. What it takes from the plan
// instead is the booked COLOUR, which is the only thing a priced slot
// tells it that a printed one cannot; see materializePlanLocked.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) priceProducedSlotsLocked(
	playerID, cardID uuid.UUID,
	slots []ProducedManaEntry,
	fromTap bool,
) []ProducedManaEntry {
	out := make([]ProducedManaEntry, 0, len(slots))
	changed := false
	for _, slot := range slots {
		amounts := make(map[string]int, len(slot.Options))
		differs := false
		for _, color := range slot.Options {
			base := slot.AmountFor(color)
			n := len(g.producedManaPreviewLocked(playerID, cardID, repeatColor(color, base), fromTap))
			amounts[color] = n
			if n != base {
				differs = true
			}
		}
		if !differs {
			out = append(out, slot)
			continue
		}
		changed = true
		if len(slot.Options) == 1 {
			color := slot.Options[0]
			for k := 0; k < amounts[color]; k++ {
				out = append(out, ProducedManaEntry{Options: []string{color}})
			}
			continue
		}
		live := ProducedManaEntry{Amounts: map[string]int{}}
		for _, color := range slot.Options {
			if amounts[color] <= 0 {
				// A colour the window produces none of is not on offer;
				// ParseProducedMana drops a zero-amount option for the
				// same reason.
				continue
			}
			live.Options = append(live.Options, color)
			live.Amounts[color] = amounts[color]
		}
		if len(live.Options) == 0 {
			continue
		}
		out = append(out, live)
	}
	if !changed {
		return slots
	}
	return out
}

// firstSlotOffering is the index of the first MULTI-OPTION slot that
// offers `color`, or oneColorSlotNone.
//
// It is how the executor finds the slot a plan booked a colour for when
// the printed slots do not show it as a one-colour pick — a Birds of
// Paradise under Mana Reflection (#1222). Multi-option only, because a
// printed single-option slot is not a pick and nothing was booked for
// it; first, because the planner refuses a source with two pickable
// slots outright (oneColorSlotUnplannable), so a plan that names a
// colour at all names it for exactly one of them.
func firstSlotOffering(slots []ProducedManaEntry, color string) int {
	if color == "" {
		return oneColorSlotNone
	}
	for i, slot := range slots {
		if len(slot.Options) > 1 && colorOffered(slot.Options, color) {
			return i
		}
	}
	return oneColorSlotNone
}

// repeatColor is n copies of one colour — the produced-mana list a slot
// of amount n adds, and the shape every caller of produceManaLocked
// builds.
func repeatColor(color string, n int) []string {
	if color == "" || n <= 0 {
		return nil
	}
	out := make([]string, n)
	for i := range out {
		out[i] = color
	}
	return out
}
