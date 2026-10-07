package game

import (
	"fmt"

	"github.com/google/uuid"
)

// target_set.go — #1559: legality of the chosen SET of targets
// (CR 601.2c), and the "X or less" bound that rides with it.
//
// Everything else about a target clause judges ONE candidate at a
// time: TargetSpec.CardOK, the zone list, the CR 702 keyword gate.
// Some printed clauses constrain the picks against each other —
// "any number of target creature cards that each have a different
// mana value" (Agadeem's Awakening), "two target creatures controlled
// by different players" (Run Away Together) — and no per-candidate
// predicate can say that, because whether a pick is legal depends on
// what else was picked.
//
// The rule is DECLARATIVE: a key read off each pick, which no two
// picks of the clause may share. Four readers answer one question
// with it, and a key is the shape all four can evaluate:
//
//   - the announce gate (validateAnnouncedTargetsWithLocked), which
//     refuses a set that breaks it with a sentence naming the rule;
//   - the CR 608.2b re-check (TargetStillLegalForEffect), which judges
//     the SURVIVING picks against it — see setRuleConflictLocked;
//   - the enumerator (internal/legal), which never builds a set that
//     breaks it;
//   - the client's picker, which greys a candidate whose key an
//     earlier pick already holds, from the map the view ships.
//
// See docs/decisions/0019-structured-targeting.md, the 2026-09-24
// amendment.

// TargetDifference is a clause's set rule: no two picks may share the
// value Key reads.
type TargetDifference struct {
	// Label is the printed rule in a form the picker and the refusal
	// can both say after "the targets must": "each have a different
	// mana value", "be controlled by different players".
	Label string

	// Key reads the value that must differ, off a card in the zone it
	// was found in. ok=false means the card has no value to compare
	// (an unreadable cost, a card the rule does not apply to), and a
	// pick with no key conflicts with nothing — the lenient reading,
	// chosen because a key the engine cannot compute is one the
	// client cannot grey either.
	//
	// Player picks have no key: no printed clause of this family
	// constrains a set of players.
	//
	// Runs under g.mu. MUST NOT call public locking mutators.
	Key func(g *Game, c Card, zone ZoneKind) (key string, ok bool)
}

// TargetSameness is the opposite set rule (#1807, ADR 0106 §5): EVERY
// pick of the clause must share the value Key names — "up to four
// target cards from a single graveyard". It is the sibling of
// TargetDifference, read by the same four readers, each of which has
// a "share" branch beside its "differ" one:
//
//   - the announce gate refuses a set whose picks hold two keys;
//   - the CR 608.2b re-check judges the SURVIVING picks, with the
//     weaker reading setRuleConflictLocked gives (two survivors that
//     disagree are both illegal);
//   - the enumerator builds sets within one key group;
//   - the client's picker greys every candidate outside the first
//     pick's group, from the map the view ships as `same`.
//
// Key is a closed vocabulary rather than a func, unlike
// TargetDifference.Key: the rule hangs off TargetSpec, which a stack
// item and a paused pick_target frame both reach, and a func there is
// a new closure route the ADR 0041 ratchet would have to carry
// (testdata/closure_fields.txt, whose ceilings may only fall). The
// only printed family — "from a single graveyard" — keys on the
// owner, so one value says it.
type TargetSameness struct {
	// Label is the printed rule in a form the picker and the refusal
	// can both say after "the targets must": "come from a single
	// graveyard".
	Label string

	// Key names the value every pick must share.
	Key TargetShareKey
}

// TargetShareKey is what a TargetSameness compares.
type TargetShareKey uint8

const (
	// TargetShareNone is the zero value: a rule with no key, which
	// constrains nothing. effects.Register refuses it on a catalog
	// clause.
	TargetShareNone TargetShareKey = iota

	// TargetShareOwner is the card's OWNER (CR 108.3). For a card in
	// a graveyard that names the graveyard it is in: CR 400.3, a card
	// that would go to a graveyard other than its owner's goes to its
	// owner's instead, so no card is ever in another player's.
	TargetShareOwner
)

// keyOf is the value a pick holds under the rule, or false when it has
// none (a card with no owner, a rule with no key) — and a pick with no
// key conflicts with nothing, TargetDifference's lenient reading.
func (s *TargetSameness) keyOf(c Card) (string, bool) {
	if s == nil {
		return "", false
	}
	switch s.Key {
	case TargetShareOwner:
		if c.Owner == uuid.Nil {
			return "", false
		}
		return c.Owner.String(), true
	}
	return "", false
}

// AnnouncedBound is what an announcement can bind an X-bounded target
// clause to (ADR 0109 §9, #1842): the X it announced (CR 107.3a — a
// mana {X}, or the count a variable sacrifice or tap names) and the
// counters its cost removed (PaidCost.CountersRemoved). A clause reads
// one of the two, by TargetSpec.BoundByCountersRemoved. A spell's
// announcement has no counter removal, and leaves the second zero.
type AnnouncedBound struct {
	X               int
	CountersRemoved int
}

// valueFor is the number the clause `s` is bounded by under this
// announcement.
func (b AnnouncedBound) valueFor(s *TargetSpec) int {
	if s.BoundByCountersRemoved {
		return b.CountersRemoved
	}
	return b.X
}

// hasXBound reports whether a clause carries an X-bound statistic
// (#1723, ADR 0109 §9): mana value X or less, mana value exactly X,
// power X or less, or toughness X or less. They share one
// bind-and-recheck mechanism and differ only in the statistic and the
// comparison boundAdmits makes.
func hasXBound(s *TargetSpec) bool {
	return s != nil && (s.ManaValueAtMostX || s.ManaValueEqualsX || s.PowerAtMostX || s.ToughnessAtMostX)
}

// boundStatistic is the statistic the clause bounds, or "" for an
// unbounded clause: "mana_value", "power" or "toughness". The wire's
// name for it, so the view and the engine say the same word.
func boundStatistic(s *TargetSpec) string {
	switch {
	case s == nil:
		return ""
	case s.ManaValueAtMostX || s.ManaValueEqualsX:
		return BoundStatManaValue
	case s.PowerAtMostX:
		return BoundStatPower
	case s.ToughnessAtMostX:
		return BoundStatToughness
	}
	return ""
}

// The statistics an X bound can read (ADR 0109 §9).
const (
	BoundStatManaValue = "mana_value"
	BoundStatPower     = "power"
	BoundStatToughness = "toughness"
)

// BoundStatisticForEffect is boundStatistic for the view: which number
// the client narrows a bounded clause's superset by. Pure.
func BoundStatisticForEffect(s *TargetSpec) string { return boundStatistic(s) }

// boundStatValue is the candidate's value of the clause's statistic,
// and false when the card has none the engine can read — an
// unreadable cost meets no mana-value bound (Card.ParsedManaValue's
// rule). Power and toughness are CR 208.1's: layers and every P/T
// counter, not clamped, the numbers PowerLE and ToughnessLE read.
func boundStatValue(g *Game, s *TargetSpec, c Card) (int, bool) {
	switch boundStatistic(s) {
	case BoundStatManaValue:
		return g.ManaValueForEffect(c)
	case BoundStatPower:
		return c.PowerForComparison(), true
	case BoundStatToughness:
		return c.CurrentToughness(), true
	}
	return 0, false
}

// boundAdmitsAt applies the clause's bound at `bound` to a candidate:
// at most `bound`, or exactly it for ManaValueEqualsX.
func boundAdmitsAt(g *Game, s *TargetSpec, c Card, bound int) bool {
	v, ok := boundStatValue(g, s, c)
	if !ok {
		return false
	}
	if s.ManaValueEqualsX {
		return v == bound
	}
	return v <= bound
}

// xBoundAdmits applies the clause's X bound to a candidate card: true
// when the clause has no bound, or the bound is not yet known (a hand
// snapshot, built before X is announced or the counters are chosen),
// or the card's statistic meets it. It reads the card with no game, so
// a spell on the stack counts {X} as zero; every engine path uses
// xBoundAdmitsIn, which counts it as the value chosen (CR 202.3e).
func (s *TargetSpec) xBoundAdmits(c Card) bool { return s.xBoundAdmitsIn(nil, c) }

// xBoundAdmitsIn is xBoundAdmits read through the game: a spell on the
// stack has the mana value its announced X gives it (CR 202.3e), which
// is what a "counter target spell with mana value X" bound (Kozilek,
// the Great Distortion) has to compare. For a card in any other zone
// the answer is the same as without the game.
func (s *TargetSpec) xBoundAdmitsIn(g *Game, c Card) bool {
	if s == nil || !hasXBound(s) || !s.xBoundSet {
		return true
	}
	return boundAdmitsAt(g, s, c, s.xBound)
}

// bindStepsBound writes the announcement's bound onto every
// X-bounded step — the announced X, or the counters removed for a
// clause that reads those. The steps hold clause COPIES
// (AnnouncedClauses), so the catalog's shared declaration is never
// touched — the same reasoning resolveStepCountsFromX gives for
// CountFromX.
func bindStepsBound(steps []AnnouncedClause, b AnnouncedBound) {
	for i := range steps {
		if c := &steps[i].Clause; hasXBound(c) {
			c.xBound, c.xBoundSet = b.valueFor(c), true
		}
	}
}

// bindStepsX is bindStepsBound for an announcement with no counter
// removal — a spell's (#1559).
func bindStepsX(steps []AnnouncedClause, x int) {
	bindStepsBound(steps, AnnouncedBound{X: x})
}

// BindStepsXForEffect is bindStepsX for the bot's enumerator, which
// builds an announcement's steps itself and has to judge its target
// sets under the X it is about to announce, exactly as the engine
// will. Pure; takes no lock.
func BindStepsXForEffect(steps []AnnouncedClause, x int) {
	bindStepsX(steps, x)
}

// StepsBoundByX reports whether any step's legality depends on the
// announcement through an X bound. Pure.
func StepsBoundByX(steps []AnnouncedClause) bool {
	for i := range steps {
		if hasXBound(&steps[i].Clause) {
			return true
		}
	}
	return false
}

// StepsBoundByCountersRemoved reports whether any step's bound reads
// the counters the activation removes rather than its X (ADR 0109
// §9). Pure.
func StepsBoundByCountersRemoved(steps []AnnouncedClause) bool {
	for i := range steps {
		if c := &steps[i].Clause; hasXBound(c) && c.BoundByCountersRemoved {
			return true
		}
	}
	return false
}

// targetDifferenceKeyLocked is the set-rule key of one pick under
// clause, or false when the clause has no rule or the pick has no
// key: a player, a card that is gone, a card the Key declines.
//
// Caller must hold g.mu.
func (g *Game) targetDifferenceKeyLocked(clause *TargetClause, ref TargetRef) (string, bool) {
	if clause == nil || clause.Different == nil || clause.Different.Key == nil {
		return "", false
	}
	return g.pickKeyLocked(ref, func(c Card, zone ZoneKind) (string, bool) {
		return clause.Different.Key(g, c, zone)
	})
}

// targetSamenessKeyLocked is targetDifferenceKeyLocked for the
// sameness rule (#1807): the key one pick holds under clause.Same, or
// false when the clause has no such rule or the pick has no key.
//
// Caller must hold g.mu.
func (g *Game) targetSamenessKeyLocked(clause *TargetClause, ref TargetRef) (string, bool) {
	if clause == nil || clause.Same == nil {
		return "", false
	}
	return g.pickKeyLocked(ref, func(c Card, _ ZoneKind) (string, bool) {
		return clause.Same.keyOf(c)
	})
}

// pickKeyLocked finds the card a pick names, in whatever zone it is in
// now, and reads a set-rule key off it. A player, or a card that is
// gone, has no key.
//
// Caller must hold g.mu.
func (g *Game) pickKeyLocked(ref TargetRef, key func(c Card, zone ZoneKind) (string, bool)) (string, bool) {
	if ref.Kind != TargetCard {
		return "", false
	}
	z := g.findCardZoneLocked(ref.ID)
	if z == nil {
		return "", false
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == ref.ID {
			return key(z.Cards[i], z.Kind)
		}
	}
	return "", false
}

// TargetDifferenceKeysForEffect is the set-rule key of each card in
// ids under spec, for the protocol projection and the enumerator. Nil
// when the spec has no rule; a card with no key is absent from the
// map. Caller must hold g.mu.
func (g *Game) TargetDifferenceKeysForEffect(spec *TargetSpec, ids []uuid.UUID) map[uuid.UUID]string {
	if spec == nil || spec.Different == nil {
		return nil
	}
	out := make(map[uuid.UUID]string, len(ids))
	for _, id := range ids {
		if k, ok := g.targetDifferenceKeyLocked(spec, TargetRef{Kind: TargetCard, ID: id}); ok {
			out[id] = k
		}
	}
	return out
}

// TargetSamenessKeysForEffect is TargetDifferenceKeysForEffect for the
// sameness rule (#1807): each card's key under spec.Same, nil when the
// spec has none. Caller must hold g.mu.
func (g *Game) TargetSamenessKeysForEffect(spec *TargetSpec, ids []uuid.UUID) map[uuid.UUID]string {
	if spec == nil || spec.Same == nil {
		return nil
	}
	out := make(map[uuid.UUID]string, len(ids))
	for _, id := range ids {
		if k, ok := g.targetSamenessKeyLocked(spec, TargetRef{Kind: TargetCard, ID: id}); ok {
			out[id] = k
		}
	}
	return out
}

// BoundValuesForEffect is each card's value of the clause's bounded
// statistic (ADR 0109 §9) — mana value, power or toughness — for the
// view to ship beside a bounded clause's legal set, so the client can
// narrow it by the X it collected or the counters it is removing. A
// card the statistic cannot be read for is absent; it meets no bound.
// Caller must hold g.mu.
func (g *Game) BoundValuesForEffect(spec *TargetSpec, ids []uuid.UUID) map[uuid.UUID]int {
	out := make(map[uuid.UUID]int, len(ids))
	for _, id := range ids {
		c := g.findCardByIDLocked(id)
		if c == nil {
			continue
		}
		if v, ok := boundStatValue(g, spec, *c); ok {
			out[id] = v
		}
	}
	return out
}

// targetSetError is the announce gate's refusal for a set that breaks
// its clause's rule. It wraps ErrIllegalTarget, so every caller that
// already maps an illegal target keeps working, and it carries the
// printed rule, so the toast says which rule was broken rather than
// "illegal target" about a set whose members are each fine.
func targetSetError(d *TargetDifference) error {
	label := "each be different"
	if d != nil && d.Label != "" {
		label = d.Label
	}
	return fmt.Errorf("%w: those targets must %s", ErrIllegalTarget, label)
}

// targetSameError is targetSetError for the sameness rule (#1807).
func targetSameError(r *TargetSameness) error {
	label := "share one key"
	if r != nil && r.Label != "" {
		label = r.Label
	}
	return fmt.Errorf("%w: those targets must %s", ErrIllegalTarget, label)
}

// setRuleConflictLocked is the CR 608.2b half of the set rule: does
// ref, which has already passed its own clause's per-candidate
// re-check, share its key with another SURVIVING pick of the same
// clause?
//
// The choice this makes, stated because CR 608.2b says a target must
// still meet the targeting requirements without saying how a
// requirement BETWEEN targets is apportioned: the rule is judged over
// the picks that are still individually legal, and when two of them
// now share a key, BOTH are illegal. Neither is preferred — the
// constraint is symmetric, and "the first one you named" is not a
// rule of the game — and it is the WEAKER of the available readings,
// never the stronger: a Run Away Together whose two creatures end up
// under one controller returns neither. A pick that became illegal on
// its own is dropped first and conflicts with nothing, so one creature
// card leaving in response to Agadeem's Awakening costs exactly that
// card.
//
// The sameness rule (#1807) is judged the same way from the other
// side: ref conflicts with a surviving pick whose key DIFFERS from its
// own, and again both are illegal. Under TargetShareOwner two
// survivors can never disagree — a card in a graveyard does not change
// owner, and one that left and came back is a new object and illegal
// on its own (CR 400.7) — so the branch is the general one, written
// for a key that could move.
//
// Caller must hold g.mu.
func (g *Game) setRuleConflictLocked(item *StackItem, clause *TargetClause, ref TargetRef) bool {
	diffKey, diffOK := g.targetDifferenceKeyLocked(clause, ref)
	sameKey, sameOK := g.targetSamenessKeyLocked(clause, ref)
	if !diffOK && !sameOK {
		return false
	}
	src := g.stackItemSourceLocked(item)
	for _, other := range item.Targets {
		if other.Kind != TargetCard || other.ID == ref.ID {
			continue
		}
		if other.Mode != ref.Mode || other.Slot != ref.Slot {
			continue
		}
		if !g.targetLegalLocked(src, clause, other) {
			continue
		}
		if diffOK {
			if k, ok := g.targetDifferenceKeyLocked(clause, other); ok && k == diffKey {
				return true
			}
		}
		if sameOK {
			if k, ok := g.targetSamenessKeyLocked(clause, other); ok && k != sameKey {
				return true
			}
		}
	}
	return false
}

// fillableCountLocked is the most picks one announcement could make
// from lt under clause: every player and card, except that cards
// sharing a set-rule key count once (#1559). "Two target creatures
// controlled by different players" with every creature on one side
// of the table has two legal candidates and no legal pair, and a
// CR 603.3d / cast-offer check that counted candidates would open a
// prompt nobody can answer.
//
// Under the sameness rule (#1807) it is the LARGEST key group, plus
// the cards with no key, which fit any group — so an exact "exile four
// target cards from a single graveyard" (Pestilent Cauldron) is not
// offered when no one graveyard holds four.
//
// Caller must hold g.mu.
func (g *Game) fillableCountLocked(clause *TargetClause, lt LegalTargets) int {
	n := len(lt.Players)
	if clause == nil || (clause.Different == nil && clause.Same == nil) {
		return n + len(lt.Cards)
	}
	if clause.Same == nil {
		return n + g.differentCountLocked(clause, lt.Cards)
	}
	groups := map[string][]uuid.UUID{}
	var order []string
	var free []uuid.UUID
	for _, id := range lt.Cards {
		k, ok := g.targetSamenessKeyLocked(clause, TargetRef{Kind: TargetCard, ID: id})
		if !ok {
			free = append(free, id)
			continue
		}
		if _, seen := groups[k]; !seen {
			order = append(order, k)
		}
		groups[k] = append(groups[k], id)
	}
	best := g.differentCountLocked(clause, free)
	for _, k := range order {
		group := append(append([]uuid.UUID(nil), free...), groups[k]...)
		if c := g.differentCountLocked(clause, group); c > best {
			best = c
		}
	}
	return n + best
}

// differentCountLocked is how many of ids one announcement could pick
// under the clause's difference rule: cards sharing a key count once,
// and a card with no key counts on its own. With no difference rule
// it is every card.
//
// Caller must hold g.mu.
func (g *Game) differentCountLocked(clause *TargetClause, ids []uuid.UUID) int {
	if clause == nil || clause.Different == nil {
		return len(ids)
	}
	n := 0
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		k, ok := g.targetDifferenceKeyLocked(clause, TargetRef{Kind: TargetCard, ID: id})
		if !ok {
			n++
			continue
		}
		if !seen[k] {
			seen[k] = true
			n++
		}
	}
	return n
}

// PickTargetSetRuleForEffect is the set rule of the clause a
// pick_target prompt is asking about, with the key of each candidate
// it offers, for the wire projection and the enumerator (#1559). Nil
// rule when the clause has none, or when the prompt carries no live
// frame (a restored snapshot), in which case the announce gate still
// enforces the rule on the answer.
//
// Caller must hold g.mu.
func (g *Game) PickTargetSetRuleForEffect(c *PendingChoice) (*TargetDifference, map[uuid.UUID]string) {
	if c == nil || c.Kind != PendingChoicePickTarget || c.pickTargetResume == nil {
		return nil, nil
	}
	clause := c.pickTargetResume.currentClause()
	if clause == nil || clause.Different == nil {
		return nil, nil
	}
	return clause.Different, g.TargetDifferenceKeysForEffect(clause, c.PickTargetCards)
}

// PickTargetSameRuleForEffect is PickTargetSetRuleForEffect for the
// sameness rule (#1807): the rule of the clause a pick_target prompt
// is asking about, with each offered card's key. Nil when the clause
// has none, or when the prompt carries no live frame (a restored
// snapshot); the announce gate still enforces the rule on the answer.
//
// Caller must hold g.mu.
func (g *Game) PickTargetSameRuleForEffect(c *PendingChoice) (*TargetSameness, map[uuid.UUID]string) {
	if c == nil || c.Kind != PendingChoicePickTarget || c.pickTargetResume == nil {
		return nil, nil
	}
	clause := c.pickTargetResume.currentClause()
	if clause == nil || clause.Same == nil {
		return nil, nil
	}
	return clause.Same, g.TargetSamenessKeysForEffect(clause, c.PickTargetCards)
}

// withoutSetRuleConflictsLocked narrows slot `slot`'s retarget
// alternatives to the cards whose set-rule key no OTHER slot of the
// same clause holds (#1559). CR 115.7c: the new target must not make
// the unchanged ones illegal, and under a set rule a shared key is
// exactly that. The gate (retargetCheckLocked) refuses such a change
// anyway; this keeps it off the offer, so a prompt never lists an
// answer the engine would refuse.
//
// Under the sameness rule (#1807) it narrows the other way: to the
// cards whose key matches the key the slots that STAY hold. That is
// the one-slot-at-a-time offer only. The gate judges the final set
// (CR 115.7e: "only the final set of targets is evaluated"), so a
// change that moves every slot to another graveyard at once — a whole
// list through RetargetStackItemForEffect — is accepted.
//
// Caller must hold g.mu.
func (g *Game) withoutSetRuleConflictsLocked(clause *TargetClause, lt LegalTargets, targets []TargetRef, slot int) LegalTargets {
	if clause == nil || (clause.Different == nil && clause.Same == nil) || slot < 0 || slot >= len(targets) {
		return lt
	}
	held := make(map[string]bool, len(targets))
	sameHeld := make(map[string]bool, len(targets))
	for i, t := range targets {
		if i == slot || t.Mode != targets[slot].Mode || t.Slot != targets[slot].Slot {
			continue
		}
		if k, ok := g.targetDifferenceKeyLocked(clause, t); ok {
			held[k] = true
		}
		if k, ok := g.targetSamenessKeyLocked(clause, t); ok {
			sameHeld[k] = true
		}
	}
	if len(held) == 0 && len(sameHeld) == 0 {
		return lt
	}
	out := LegalTargets{Players: lt.Players}
	for _, id := range lt.Cards {
		ref := TargetRef{Kind: TargetCard, ID: id}
		if k, ok := g.targetDifferenceKeyLocked(clause, ref); ok && held[k] {
			continue
		}
		// A card with a key fits only when every staying slot holds
		// that same key; staying slots that already disagree (left
		// alone under CR 115.7d) admit no keyed card at all.
		if k, ok := g.targetSamenessKeyLocked(clause, ref); ok && len(sameHeld) > 0 && (len(sameHeld) > 1 || !sameHeld[k]) {
			continue
		}
		out.Cards = append(out.Cards, id)
	}
	return out
}

// TargetsWithinXForEffect reports whether every pick that answers an
// X-bounded step meets that bound at x (#1559, #1723) — the
// enumerator's re-check for a set it priced at an X other than the one
// it bound the steps to. Picks of other steps pass. An announcement
// with no counter removal; see TargetsWithinBoundForEffect.
//
// Caller must hold g.mu.
func (g *Game) TargetsWithinXForEffect(steps []AnnouncedClause, targets []TargetRef, x int) bool {
	return g.TargetsWithinBoundForEffect(steps, targets, AnnouncedBound{X: x})
}

// TargetsWithinBoundForEffect reports whether every pick that answers
// an X-bounded step meets that step's bound under `b` — the announced
// X, or the counters removed, whichever the clause reads (ADR 0109
// §9). The enumerator's filter for a set it built against the unbound
// superset: one per counter payment for Simic Manipulator, one per
// sacrifice or tap count for Ruthless Technomancer and Aryel. Picks of
// other steps pass.
//
// Caller must hold g.mu.
func (g *Game) TargetsWithinBoundForEffect(steps []AnnouncedClause, targets []TargetRef, b AnnouncedBound) bool {
	for _, t := range targets {
		if t.Kind != TargetCard {
			continue
		}
		for i := range steps {
			clause := &steps[i].Clause
			if steps[i].Mode != t.Mode || steps[i].Slot != t.Slot || !hasXBound(clause) {
				continue
			}
			c := g.findCardByIDLocked(t.ID)
			if c == nil || !boundAdmitsAt(g, clause, *c, b.valueFor(clause)) {
				return false
			}
		}
	}
	return true
}
