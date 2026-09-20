package game

import "github.com/google/uuid"

// tap_others_cost.go — #758: tapping permanents OTHER than the source
// as part of an ability's cost (CR 602.2, CR 118.3).
//
// This is the sixth kind of cost the engine models and the second
// that taps. It is deliberately NOT tap_cost.go's TapPermanentsCost,
// and the difference is the one that matters:
//
//	TapPermanentsCost (S22, convoke / waterbend) answers "what MAY be
//	spent against a mana demand this cast already has". It is
//	optional, its size is bounded by the cost, and tapping nothing is
//	a legal cast.
//
//	TapOthersCost (#758) answers "what does this ACTIVATION owe". It
//	is a component of AbilityCost in the sense ADR 0020 §2 means:
//	a fixed number of permanents must be tapped or the ability cannot
//	be activated at all (CR 118.3).
//
// Earthcraft, Opposition, Azami, Springleaf Drum, Heritage Druid and
// the station ability (CR 702.184a) all print it. ADR 0071 §2 names
// this field — AbilityCost.TapOthers — so that effects.Station() can
// be written the day it lands.
//
// Three rules the shape exists to get right:
//
//   - **It is not the {T} symbol.** CR 302.6 and CR 602.5a hang
//     summoning sickness on {T} and {Q}, so a creature that arrived
//     this turn may still be tapped to pay this cost. There is
//     deliberately no sickness check below — the same deliberate
//     absence tap_cost.go carries for convoke, and the same one crew
//     carries (CR 702.122a).
//
//   - **"Another" is a property of the printed cost, not of the
//     engine.** Earthcraft taps "an untapped creature you control";
//     station taps "ANOTHER untapped creature you control"; Azami may
//     tap herself. ExcludeSource carries the difference, so neither
//     reading is hard-coded.
//
//   - **An already-tapped permanent cannot pay** (CR 118.3): you
//     cannot tap what is already tapped, so a board that cannot
//     produce Count untapped matches cannot activate the ability.
//
// And one it shares with every other cost component: paying a cost
// does not TARGET (CR 601.2h / 602.2b), so the filter is matched with
// specMatchLocked(…, false) and shroud, hexproof and protection never
// apply — exactly as counter_cost.go and activated.go's SacrificeOther
// match theirs.
//
// Scope of this file. It is the whole component EXCEPT its two lines
// on AbilityCost: the field itself, and the validate/pay calls in the
// activation path. Those land with the wire and the client in the
// next slice of #758, and they are one call per component because
// this file keeps the same entry-point shape counter_cost.go
// established — one options walk, one payability predicate, one
// validator, one payer, shared by the engine, the enumerator and the
// view so they cannot disagree (#544).

// TapOthersCost is the "tap N untapped [permanents] you control"
// component of an ability's cost, named by ADR 0071 §2.
//
// The zero value, and nil, demand nothing — so the activation path
// can ask without a guard, the way it asks TapPermanentsCost.
type TapOthersCost struct {
	// Count is how many permanents must be tapped. 1 for station,
	// Earthcraft and Springleaf Drum; 2 for Clock of Omens; 3 for
	// Heritage Druid; 4 for Nullmage Shepherd.
	//
	// Fixed. A VARIABLE count ("tap X untapped Foods you control",
	// Apothecary White) is not modelled here and is not a matter of
	// reading this field differently: it would need the announce
	// path MinX uses, and it is stated out of scope on #758.
	Count int

	// Filter is what may be tapped ("an untapped creature you
	// control", "an artifact you control"), reused from the
	// TargetSpec vocabulary exactly as AbilityCost.SacrificeOther and
	// CounterRemovalCost.From reuse it.
	//
	// It does NOT target. The "you control" clause and the untapped
	// clause are enforced by the validator below rather than asked of
	// the spec, so a filter that forgot either cannot let one
	// permanent pay twice or let an opponent's creature pay at all.
	//
	// Nil is not a cost: Empty reports true, because "tap nothing in
	// particular" is not a printed clause.
	Filter *TargetSpec

	// ExcludeSource is the printed word "another" (CR 702.184a's
	// station, Jaspera Sentinel's "{T}, tap another untapped creature
	// you control"). When set, the ability's own source may not be
	// named.
	//
	// When it is CLEAR the source may pay — Azami, Lady of Scrolls
	// taps a Wizard you control and is herself a Wizard. The
	// interaction with AbilityCost.Tap is the validator's, not this
	// field's: a cost that also taps the source has already spent it,
	// and the source is then unavailable because it is tapped, not
	// because of a rule of its own.
	ExcludeSource bool

	// Label is the clause as printed, shown above the client's
	// picker so the prompt reads like the card. "Tap three untapped
	// Elves you control", not "choose 3".
	Label string
}

// Empty reports whether the cost demands nothing. Nil-safe, so the
// activation path can ask without a guard.
func (c *TapOthersCost) Empty() bool {
	return c == nil || c.Count < 1 || c.Filter == nil
}

// TapOthersOptionsForEffect is the set of permanents that could pay
// `tc` for an activation of an ability on `sourceID` by `playerID`,
// in battlefield order. Empty when nothing on the board matches.
//
// ONE walk, shared by the protocol view (the client's picker), the
// legal-move enumerator (the bots) and — through the same predicates
// — validateTapOthersCostLocked, so the three cannot disagree about
// which permanent pays. That is the #544 invariant and the reason
// counter_cost.go has exactly one of these too.
//
// It is the NON-targeting candidate walk (specCandidatesLocked):
// choosing a permanent to pay a cost does not target it, so a
// hexproof creature is still a legal tap.
//
// Whether the options add up to Count is the PAYMENT's question,
// answered by TapOthersPayable and by the validator. This walk never
// enumerates subsets — that is the combinatorial expansion #544
// warns about.
//
// Caller must hold g.mu (read or write).
func (g *Game) TapOthersOptionsForEffect(playerID, sourceID uuid.UUID, tc *TapOthersCost) []uuid.UUID {
	if tc.Empty() {
		return nil
	}
	var out []uuid.UUID
	for _, id := range g.specCandidatesLocked(playerID, tc.Filter).Cards {
		if tc.ExcludeSource && id == sourceID {
			continue
		}
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != playerID || c.Tapped {
			continue
		}
		out = append(out, id)
	}
	return out
}

// TapOthersPayable reports whether `tc` could be paid at all right
// now — the question the legal enumerator and the client's greyed
// menu row ask before offering the activation, and the one CR 118.3
// turns into a hard refusal.
//
// Caller must hold g.mu (read or write).
func (g *Game) TapOthersPayable(playerID, sourceID uuid.UUID, tc *TapOthersCost) bool {
	if tc.Empty() {
		return true
	}
	return len(g.TapOthersOptionsForEffect(playerID, sourceID, tc)) >= tc.Count
}

// validateTapOthersCostLocked checks that every permanent the
// activator named may actually be tapped for this cost, without
// tapping anything — ADR 0020 §3's "validate everything, then pay
// everything", so a refused activation never leaves a board half
// tapped.
//
// What it enforces, in the order the errors matter:
//
//   - IDs sent for an ability with no tap-others component are
//     rejected rather than ignored, as crew_ids, sacrifice_ids and
//     counter_source_ids are: a client that sends them is confused
//     about which ability it is firing.
//   - Exactly Count permanents, no more and no fewer. A cost is not
//     partially payable (CR 118.3), and over-naming would tap a
//     permanent for nothing.
//   - No permanent named twice — otherwise one creature pays both
//     halves of "tap two untapped artifacts you control".
//   - The source itself when ExcludeSource is set ("another").
//   - On the battlefield (ErrCardNotFound), controlled by the
//     activator (ErrCardCallerMismatch), untapped (ErrAlreadyTapped)
//     and matched by Filter WITHOUT the targeting gate
//     (ErrIllegalTarget).
//
// There is no summoning-sickness check, deliberately: this is not
// the {T} symbol (CR 302.6, CR 602.5a).
//
// Caller must hold g.mu.
func (g *Game) validateTapOthersCostLocked(playerID, sourceID uuid.UUID, tc *TapOthersCost, ids []uuid.UUID) error {
	if tc.Empty() {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(ids) != tc.Count {
		return ErrInvalidParam
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		if tc.ExcludeSource && id == sourceID {
			return ErrInvalidParam
		}
		c := findBattlefieldCard(g, id)
		if c == nil {
			return ErrCardNotFound
		}
		// "You control" is a cost's own clause, enforced here rather
		// than asked of the filter, so a filter that forgot it cannot
		// spend an opponent's board.
		if c.Controller != playerID {
			return ErrCardCallerMismatch
		}
		// CR 118.3: you can't tap what is already tapped, so the
		// cost is simply unpayable with this permanent.
		if c.Tapped {
			return ErrAlreadyTapped
		}
		// specMatchLocked(…, false), not targetLegalLocked: tapping a
		// permanent to pay a cost does not target it (CR 601.2h /
		// 602.2b), so the CR 702 keyword gate must not apply.
		if !g.specMatchLocked(SourceChooser(playerID), tc.Filter, TargetRef{Kind: TargetCard, ID: id}, false) {
			return ErrIllegalTarget
		}
	}
	return nil
}

// payTapOthersCostLocked taps the validated permanents, emitting one
// EventTapCard each so that "whenever a permanent becomes tapped"
// sees every one of them, and returns the cards AS THEY WERE TAPPED.
//
// The snapshot is the point of the return value, not a convenience.
// A station ability puts charge counters equal to the tapped
// creature's power (CR 702.184a) and crew reads power the same way
// (CR 702.122a): the number is fixed at PAYMENT time, so the tapped
// creature dying, shrinking or leaving before the ability resolves
// changes nothing. The caller records the number on the stack item,
// never the creature's ID.
//
// Call only after validateTapOthersCostLocked has passed, and — as
// with payAdditionalCostLocked and payTapPermanentsCostLocked — after
// the ability is on the stack, so a tap-watching trigger goes above
// it and resolves first (Opposition into a "becomes tapped" payoff).
//
// A permanent that vanished between validation and payment is
// skipped rather than erroring: nothing can happen between the two
// under one lock today, and a partial board is still better than a
// panic if that ever stops being true.
//
// Caller must hold g.mu.
func (g *Game) payTapOthersCostLocked(playerID uuid.UUID, ids []uuid.UUID) []Card {
	paid := make([]Card, 0, len(ids))
	for _, id := range ids {
		c := findBattlefieldCard(g, id)
		if c == nil || c.Tapped {
			continue
		}
		c.Tapped = true
		paid = append(paid, *c)
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: id})
	}
	return paid
}
