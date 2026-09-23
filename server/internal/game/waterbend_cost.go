package game

import "github.com/google/uuid"

// waterbend_cost.go — #1310 / #1311: waterbend (CR 701.67) as a cost
// that is NOT a spell's. S22 built it for a cast (`Spec.TapCost`,
// tap_cost.go); this file is the same component on the two other
// surfaces that print it:
//
//	an activated ability   "Waterbend {8}: Transform Aang."
//	                        (Aang, Swift Savior; Katara, Water Tribe's
//	                        Hope; Avatar Kuruk) — AbilityCost.Waterbend
//	a pay-or-counter prompt "Ward—Waterbend {4}" (The Unagi of Kyoshi
//	                        Island) — CounterUnlessPaidPrompt.Waterbend
//
// CR 701.67a: "Waterbend [cost]" means "Pay [cost]. For each generic
// mana in that cost, you may tap an untapped artifact or creature you
// control rather than pay that mana." So the payment is MANA with a
// second way to cover each generic symbol, and the struct that says so
// is tap_cost.go's TapPermanentsCost — one component, now with three
// owners. Nothing here is a new kind of cost.
//
// ONE difference from the cast side, and it is where the mana LIVES.
// On a spell the waterbend cost is an ADDITIONAL cost ("as an
// additional cost to cast this spell, waterbend {X}"), so
// TapPermanentsCost.Extra is added on top of the printed cost. On an
// ability the waterbend cost IS the cost — "Waterbend {8}:" has no
// other mana — so the waterbend mana is already inside
// AbilityCost.Mana, and Extra names which PART of that mana the taps
// may cover (CR 701.67b: only the generic in the waterbend cost, never
// the rest of the total). Keeping the mana in AbilityCost.Mana is what
// lets every existing reader of an ability's mana — X detection, the
// CR 601.2f cost-modifier pass (Boom Scholar reaches Kuruk's exhaust
// ability), the view's cost chip, the enumerator's affordability —
// keep working without learning that waterbend exists.
//
// And one rule both owners share with #758's TapOthersCost: tapping to
// pay a cost is NOT the {T} symbol, so summoning sickness does not
// apply (CR 302.6) and a creature that arrived this turn may waterbend.
// Aang may even tap himself to pay for his own transform.

// WaterbendBudget is how many permanents may be tapped against a
// waterbend clause, given the mana the cost will actually charge.
//
// The clause's own generic (plus its X slots at the announced X) is
// the printed ceiling — CR 701.67b. The priced cost's generic is the
// practical one: a CR 601.2f discount that has already removed some of
// the generic leaves fewer symbols for a tap to pay, and tapping a
// creature for a symbol that no longer exists would tap it for
// nothing. The smaller of the two is the answer.
//
// Pure — no game state. `priced` is AbilityManaCostForEffect's answer
// for an ability, or the parsed prompt cost for a pay-unless. Zero for
// an empty clause.
func WaterbendBudget(wb *TapPermanentsCost, priced ParsedCost, x int) int {
	if wb.Empty() || wb.Extra == "" {
		return 0
	}
	clause, err := ParseCost(wb.Extra)
	if err != nil {
		return 0
	}
	budget := clause.Generic + clause.XSlots*x
	if owed := priced.Generic + priced.XSlots*x; owed < budget {
		budget = owed
	}
	if budget < 0 {
		return 0
	}
	return budget
}

// WaterbendReduced is `priced` with `n` generic symbols paid by
// tapping instead of by mana. The {X} is folded into Generic at the
// announced X first, so the subtraction has one concrete number to
// work against; the solver computes Generic + XSlots*x anyway, so the
// fold changes nothing it reads.
//
// Pure. `n` is expected to be within WaterbendBudget — the validator
// has refused anything larger — and is clamped at the generic anyway,
// because a cost with negative generic is not a cost.
func WaterbendReduced(priced ParsedCost, x, n int) ParsedCost {
	if n <= 0 {
		return priced
	}
	out := priced
	out.Required = append([]ColorRequirement(nil), priced.Required...)
	out.Generic += out.XSlots * x
	out.XSlots = 0
	if n > out.Generic {
		n = out.Generic
	}
	out.Generic -= n
	return out
}

// WaterbendOptionsForEffect is the set of permanents `playerID` could
// tap against `wb` right now, in battlefield order: untapped, under
// their control, and matched by the clause (an artifact or a creature).
//
// ONE walk for the engine's validator, the protocol view's picker and
// the legal-move enumerator, so the three cannot disagree about which
// permanent may pay (#544). It is the NON-targeting candidate walk:
// tapping a permanent to pay a cost does not target it (CR 601.2h /
// 602.2b), so a hexproof creature of yours is a legal tap.
//
// `exclude` is a permanent that is already paying another component of
// the same cost — the ability's own source when its cost also prints
// {T}. uuid.Nil excludes nothing.
//
// Caller must hold g.mu (read or write).
func (g *Game) WaterbendOptionsForEffect(playerID, exclude uuid.UUID, wb *TapPermanentsCost) []uuid.UUID {
	if wb.Empty() {
		return nil
	}
	var out []uuid.UUID
	for _, id := range g.specCandidatesLocked(playerID, wb.Spec).Cards {
		if id == exclude {
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

// AbilityWaterbendExclusion is the permanent an activation's waterbend
// taps may not name because another component of the same cost already
// spends it: the source, when the cost also prints {T} (CR 118.3 — it
// cannot be tapped twice). uuid.Nil otherwise, and in particular for
// Aang, whose "Waterbend {8}" has no {T} and who may help pay for his
// own transform.
func AbilityWaterbendExclusion(sourceID uuid.UUID, cost AbilityCost) uuid.UUID {
	if cost.Tap {
		return sourceID
	}
	return uuid.Nil
}

// validateAbilityWaterbendLocked checks an activation's waterbend taps
// without tapping anything — ADR 0020 §3's "validate everything, then
// pay everything".
//
// The per-permanent rules are tap_cost.go's validator verbatim (on the
// battlefield, controlled by the activator, untapped, matched by the
// clause without the targeting gate, named once, no more than the
// budget), so a cast and an activation cannot disagree about who may
// waterbend. What is added here is what only the activation path can
// see, because it holds every component of the one cost:
//
//   - the source when the cost also prints {T};
//   - a permanent already named to another component of the SAME cost.
//     For a crew tap or a tap-another tap that is CR 118.3 outright —
//     a tapped permanent cannot be tapped again. For a sacrifice or a
//     return it is the engine's payment order: those move the
//     permanent before the taps are paid (the taps go after the
//     ability is on the stack), so the discount would be granted and
//     the tap that bought it would never happen. Paper lets you tap
//     then sacrifice; no printed card combines the two, and refusing
//     is the weaker-than-printed direction (#259).
//
// Caller must hold g.mu.
func (g *Game) validateAbilityWaterbendLocked(playerID, sourceID uuid.UUID, cost AbilityCost, params ActivateAbilityParams, budget int) error {
	if err := g.validateTapPermanentsCostLocked(playerID, cost.Waterbend, params.WaterbendIDs, budget); err != nil {
		return err
	}
	if len(params.WaterbendIDs) == 0 {
		return nil
	}
	taken := make(map[uuid.UUID]bool)
	if ex := AbilityWaterbendExclusion(sourceID, cost); ex != uuid.Nil {
		taken[ex] = true
	}
	for _, l := range [][]uuid.UUID{params.CrewIDs, params.TapIDs, params.SacrificeIDs, params.ReturnIDs} {
		for _, id := range l {
			taken[id] = true
		}
	}
	for _, id := range params.WaterbendIDs {
		if taken[id] {
			return ErrInvalidParam
		}
	}
	return nil
}
