package game

import "github.com/google/uuid"

// next_spell_promise.go — #1852: "the next <kind> spell you cast this
// turn <does something>".
//
//	Printed example                         Riders
//	Savage Summoning   next creature spell  flash + can't be countered + an extra +1/+1 counter
//	Quicken            next sorcery spell   flash
//	Hardened Berserker next spell           costs {1} less
//	Kaza, Roil Chaser  next instant/sorcery costs {X} less
//
// ONE PROMISE, ONE SPEND. The promise is held on the caster
// (Player.Statics, PlayerStatic.NextSpell) and is decided exactly once,
// at the moment a spell becomes cast (CR 601.2i): castSpellLocked calls
// spendNextSpellPromisesLocked beside the can't-be-countered spend. The
// first spell the filter matches takes it, whether or not that spell is
// later countered (the rulings are explicit that a countered spell
// still used it up), and a spell the filter does not match leaves it
// for the next one. A copy is not cast (CR 707.10) and a land is not
// cast (CR 305.1), so neither ever reaches the spend.
//
// THE RIDERS ARE READ AT THREE DIFFERENT TIMES, which is why the
// promise is data rather than a callback:
//
//   - Flash is read BEFORE the cast, because it decides whether the
//     cast may begin at all. castTimingVerdictLocked folds a live,
//     matching promise into the same verdict a per-player timing
//     statement does, so CastSpell, the bot enumerator and the view's
//     castable_here all agree through CastTimingOpenLocked.
//   - The cost reduction is read while the cast is PRICED (CR 601.2f).
//     activeCostModifiersLocked binds a live matching promise as one
//     more reduction, so the payment path, the auto-tap preview and the
//     enumerator price the same cast the same way. It is consumed at the
//     spend, so a cast that fails to pay does not use it up.
//   - The extra counter and the uncounterable rider are read AFTER the
//     cast, as marks the spend writes onto the spell's stack item:
//     StackItem.PromisedCounters (read by applyCastEntryCountersLocked,
//     so Doubling Season sees the counter) and the existing
//     StackItem.CantBeCountered marks (read by the one counter gate).
//     A mark lives on the OBJECT, so it ends with it (CR 400.7) and a
//     copy never has it.
//
// A card whose promise needs a rider none of those cover registers a
// body with RegisterCastFollowUp and names it in FollowUp; it runs once
// the spell is on the stack and EventCast is out, exactly as a cast
// permission's follow-up does.
//
// PLAIN DATA throughout: the promise is scalars (so clonePlayerStatics'
// value copy is a deep copy) and the mark is a small struct, so a table
// holding either is still a restore point.

// NextSpellPromise is the NextSpell payload of a PlayerStatic: one
// "the next <kind> spell you cast" promise and the riders it carries.
type NextSpellPromise struct {
	// Active is the presence bit. The zero filter is "every spell",
	// which is Hardened Berserker's real promise, so the payload needs
	// a bit of its own (CounterShieldGrant.Active's argument).
	Active bool `json:"active,omitempty"`

	// Filter narrows which spells the promise is about: CreatureOnly
	// (Savage Summoning), SorceryOnly (Quicken), InstantOrSorceryOnly
	// (Kaza), NoncreatureOnly. The zero filter is every spell.
	//
	// No `omitzero`, for PlayerStatic.Timing's reason (#1492).
	Filter PermissionFilter `json:"filter"`

	// Flash is "can be cast as though it had flash" (CR 702.8).
	Flash bool `json:"flash,omitempty"`

	// Reduce is "costs {N} less to cast": generic mana, floored at the
	// spell's generic component like every CR 601.2f reduction. A card
	// that scales (Kaza's Wizards) computes N when it resolves and
	// stores the number, which is what the printed text says.
	Reduce int `json:"reduce,omitempty"`

	// AffinityFor is "has affinity for <type>" (CR 702.41a): the spell
	// costs {1} less for each permanent of that card type the caster
	// controls as they cast it (Saheeli, the Gifted's "artifacts").
	// Counted when the spell is priced, in the same CR 601.2f reduction
	// as Reduce, so tapping an artifact for mana still counted it.
	AffinityFor string `json:"affinityFor,omitempty"`

	// CounterKind and Counters are "that creature enters with N
	// additional <kind> counters on it". Zero Counters is no rider.
	CounterKind string `json:"counterKind,omitempty"`
	Counters    int    `json:"counters,omitempty"`

	// CantBeCountered is "that spell can't be countered".
	CantBeCountered bool `json:"cantBeCountered,omitempty"`

	// FollowUp names a body registered with RegisterCastFollowUp, run
	// when the promise is spent.
	FollowUp CastFollowUpKey `json:"followUp,omitempty"`

	// EveryMatchingSpell makes the promise about EVERY matching spell
	// cast while it lasts rather than the next one: it is never spent,
	// and it ends with its duration. "Artifact, instant, and sorcery
	// spells you cast this turn cost {2} less to cast" (Urza,
	// Planeswalker; ADR 0145). Only the price rider (Reduce) is read for
	// one; the other riders describe a single spell.
	EveryMatchingSpell bool `json:"everyMatchingSpell,omitempty"`

	// WithoutPayingManaCost is "can be cast without paying its mana
	// cost" (Apex Observatory, #2709): while the promise stands, a
	// matching spell is offered one more alternative cost (CR 118.9),
	// free, under GrantedAltCostNextFree — read beside the permanents'
	// granted offers (grantedAlternativeCostsLocked), so the view, the
	// strip, the enumerator and CastSpell all see it. The promise is
	// spent by the next matching spell whichever cost it was cast for:
	// that spell IS the next one.
	WithoutPayingManaCost bool `json:"withoutPayingManaCost,omitempty"`

	// Text is the clause as printed, for the log and the panel. Not a
	// rules input.
	Text string `json:"text,omitempty"`
}

// PromisedCounter is one "enters with an additional counter" mark on a
// stack item (StackItem.PromisedCounters), written when a promise is
// spent on the spell and read at the spell's entry.
type PromisedCounter struct {
	Kind       string    `json:"kind"`
	Count      int       `json:"count"`
	Source     uuid.UUID `json:"source,omitempty"`
	SourceName string    `json:"sourceName,omitempty"`
}

// copyPromisedCounters gives a mark slice its own backing array. Nil
// for none, so an unmarked item stays the zero value it always was.
func copyPromisedCounters(in []PromisedCounter) []PromisedCounter {
	if len(in) == 0 {
		return nil
	}
	return append([]PromisedCounter(nil), in...)
}

// GrantNextSpellPromiseForEffect gives `player` a one-use promise about
// the next spell they cast. `d` is how long it waits: the turn
// (g.UntilEndOfTurnDuration(), every "this turn" card) or
// IndefiniteDuration() for one with no stated end (Xho Cai). A zero
// duration is read as until end of turn, the narrowest real one.
//
// The *ForEffect surface: caller must hold g.mu (write).
func (g *Game) GrantNextSpellPromiseForEffect(player uuid.UUID, promise NextSpellPromise, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	promise.Active = true
	p.Statics = append(p.Statics, PlayerStatic{
		NextSpell: promise,
		Source:    source,
		Label:     label,
		Duration:  d,
	})
}

// livePromise reports whether a stored entry is a promise that still
// stands. The duration is tested here as well as in the sweep, for
// playerAbilityTokensLocked's reason: the sweep is hygiene, the reader
// is the truth.
func (g *Game) livePromise(s PlayerStatic) bool {
	return s.NextSpell.Active && !g.durationExpiredLocked(s.Duration, false)
}

// nextSpellCostModifiersLocked binds every live promise of `player`'s
// that reduces the price of `card` as an ordinary CR 601.2f reduction,
// so it composes with every other modifier and is clamped like them.
// Built fresh per query from data; nothing is stored.
//
// Caller must hold g.mu (read or write).
func (g *Game) nextSpellCostModifiersLocked(player uuid.UUID, card Card) []boundCostModifier {
	p := g.playerByIDLocked(player)
	if p == nil {
		return nil
	}
	var out []boundCostModifier
	for _, s := range p.Statics {
		np := s.NextSpell
		if (np.Reduce <= 0 && np.AffinityFor == "") || !g.livePromise(s) || !np.Filter.Matches(card) {
			continue
		}
		n := np.Reduce + g.affinityCountLocked(player, np.AffinityFor)
		if n <= 0 {
			continue
		}
		src := card
		src.Controller = player
		out = append(out, boundCostModifier{
			modifier: CostModifier{
				Kind:   CostReduction,
				Label:  np.Text,
				Amount: func(CostQuery) int { return n },
			},
			source: src,
		})
	}
	return out
}

// affinityCountLocked is the number of permanents of card type
// `cardType` that `player` controls right now, read from their
// effective types (CR 702.41a). Zero for no affinity.
//
// Caller must hold g.mu (read or write).
func (g *Game) affinityCountLocked(player uuid.UUID, cardType string) int {
	if cardType == "" || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && characteristicHasType(SourceCharacteristics(c), cardType) {
			n++
		}
	}
	return n
}

// spendNextSpellPromisesLocked is CR 601.2i for the one-use promise:
// the spell `spellID` has just become cast by `caster`, so every live
// promise of the caster's whose filter it matches is taken off
// Player.Statics and its riders land on the spell's item. A spell that
// matches none changes nothing. It returns the follow-up keys of the
// promises it spent, for castSpellLocked to run once EventCast is out.
//
// Replaces the statics slice rather than compacting it, and the mark
// slices rather than appending into them, for the reason
// sweepPlayerStaticsLocked gives: an undo snapshot shares the arrays.
//
// Caller must hold g.mu (write).
func (g *Game) spendNextSpellPromisesLocked(caster, spellID uuid.UUID) []CastFollowUpKey {
	p := g.playerByIDLocked(caster)
	item := g.StackMeta[spellID]
	spell := g.spellCardOnStackLocked(spellID)
	if p == nil || item == nil || spell == nil || item.IsCopy {
		return nil
	}
	var (
		kept      []PlayerStatic
		spent     []PlayerStatic
		followUps []CastFollowUpKey
	)
	for _, s := range p.Statics {
		if g.livePromise(s) && !s.NextSpell.EveryMatchingSpell && s.NextSpell.Filter.Matches(*spell) {
			spent = append(spent, s)
			continue
		}
		kept = append(kept, s)
	}
	if len(spent) == 0 {
		return nil
	}
	p.Statics = kept
	for _, s := range spent {
		np := s.NextSpell
		if np.CantBeCountered {
			item.CantBeCountered = append(copyCounterShieldMarks(item.CantBeCountered),
				CounterShieldMark{Source: s.Source, SourceName: s.Label, Label: np.Text})
		}
		if np.Counters > 0 && np.CounterKind != "" {
			item.PromisedCounters = append(copyPromisedCounters(item.PromisedCounters),
				PromisedCounter{Kind: np.CounterKind, Count: np.Counters, Source: s.Source, SourceName: s.Label})
		}
		if np.FollowUp != "" {
			followUps = append(followUps, np.FollowUp)
		}
	}
	return followUps
}

// promisedEntryCountersLocked seeds the counters a spent promise marked
// onto the spell's item into the entry event, beside the mana-rider and
// printed "enters with" clauses, so the CR 616 pipeline (Doubling
// Season, Hardened Scales) sees them.
//
// Caller must hold g.mu.
func (g *Game) promisedEntryCountersLocked(ev *ReplacementEvent, item *StackItem) {
	if ev == nil || item == nil {
		return
	}
	for _, m := range item.PromisedCounters {
		if m.Count > 0 && m.Kind != "" {
			ev.AddCounterAtETB(m.Kind, m.Count)
		}
	}
}
