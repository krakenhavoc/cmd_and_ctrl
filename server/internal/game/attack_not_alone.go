package game

import (
	"strings"

	"github.com/google/uuid"
)

// attack_not_alone.go is "this creature can't attack alone" (Pet
// Sehlat, #4973): the AttackLimitNotAlone scope of attack_limits.go.
//
// CR 506.5: "A creature attacks alone if it's the only creature declared
// as an attacker during the declare attackers step." CR 508.1c makes the
// restriction a property of the declaration, and its example says two
// creatures that each "can't attack alone" may attack together. So
// nothing here forbids the creature; it forbids a DECLARATION whose one
// attacker is such a creature. See the scope's comment in
// attack_limits.go for where that is judged and why in two places.
//
// #6197 widens "alone" to a companion rule: Orcish Conscripts "can't
// attack unless at least two other creatures attack" and Okk "can't
// attack unless a creature with greater power also attacks". The
// judgement is the same, read as "fewer companions than the rule needs"
// instead of "the only attacker"; "can't attack alone" is the rule that
// needs one companion of any power.

// companionRule is what one NotAlone limit asks of a declaration that
// has its source in it (#6197): `need` other attackers, each with power
// greater than the source's when `greater` is set. Pet Sehlat's "can't
// attack alone" is need 1; Orcish Conscripts' "unless at least two
// other creatures attack" is need 2; Okk's "unless a creature with
// greater power also attacks" is need 1, greater.
type companionRule struct {
	need    int
	greater bool
	// colors, when set, narrows a companion to a creature of one of these
	// colors (Scarred Puma's "a black or green creature also attacks").
	colors []string
}

// tighter is the stricter of two rules on one creature: it must meet
// both, so the larger count and either power condition.
//
// One creature never has two different color conditions in the catalog, so
// the first rule that names colors is the one kept.
func (r companionRule) tighter(o companionRule) companionRule {
	colors := r.colors
	if len(colors) == 0 {
		colors = o.colors
	}
	return companionRule{need: max(r.need, o.need), greater: r.greater || o.greater, colors: colors}
}

// countsAsCompanion reports whether `other` keeps `src` company under
// `r`: any other creature, or, for Okk, one with greater power, read at
// declaration (CR 508.1c).
func (r companionRule) countsAsCompanion(src, other *Card) bool {
	if other == nil || src == nil || other.InstanceID == src.InstanceID {
		return false
	}
	if r.greater && other.PowerForComparison() <= src.PowerForComparison() {
		return false
	}
	if len(r.colors) == 0 {
		return true
	}
	for _, have := range other.EffectiveColors() {
		for _, want := range r.colors {
			if strings.EqualFold(have, want) {
				return true
			}
		}
	}
	return false
}

// notAloneSourcesLocked is every creature an active "can't attack
// alone" limit is about, with its companion rule, or nil — the answer
// at nearly every table.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) notAloneSourcesLocked() map[uuid.UUID]companionRule {
	var out map[uuid.UUID]companionRule
	for _, l := range g.activeAttackLimitsLocked() {
		if l.limit.Scope != AttackLimitNotAlone || l.source == nil {
			continue
		}
		if out == nil {
			out = map[uuid.UUID]companionRule{}
		}
		r := companionRule{need: l.limit.companionsNeeded(), greater: l.limit.OthersPowerGreater, colors: l.limit.OthersColors}
		if prev, ok := out[l.source.InstanceID]; ok {
			r = r.tighter(prev)
		}
		out[l.source.InstanceID] = r
	}
	return out
}

// companionsInLocked counts the attackers in `assign` that keep `src`
// company under `r`.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) companionsInLocked(src *Card, r companionRule, assign map[uuid.UUID]uuid.UUID) int {
	n := 0
	for id, t := range assign {
		if t == uuid.Nil || id == src.InstanceID {
			continue
		}
		if r.countsAsCompanion(src, findBattlefieldCard(g, id)) {
			n++
		}
	}
	return n
}

// unmetNotAloneAttackerLocked is the first creature in `assign`, in
// battlefield order, whose companion rule `assign` does not meet, with
// that rule, or nil. For "can't attack alone" that is the one attacker
// of a one-creature declaration.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) unmetNotAloneAttackerLocked(assign map[uuid.UUID]uuid.UUID, rules map[uuid.UUID]companionRule) (*Card, companionRule) {
	if len(rules) == 0 || len(assign) == 0 || g.Battlefield == nil {
		return nil, companionRule{}
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		r, ok := rules[c.InstanceID]
		if !ok || assign[c.InstanceID] == uuid.Nil {
			continue
		}
		if g.companionsInLocked(c, r, assign) < r.need {
			return c, r
		}
	}
	return nil, companionRule{}
}

// companionsReachableLocked reports whether enough creatures could be
// added to `assign` to meet `src`'s rule: the companions already in it
// plus the distinct creatures attackCompanionsLocked offers that count
// under `r`. With `free`, only companions that attack at no CR 508.1h
// price count.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) companionsReachableLocked(src *Card, r companionRule, assign map[uuid.UUID]uuid.UUID, free bool) bool {
	have := g.companionsInLocked(src, r, assign)
	if have >= r.need {
		return true
	}
	seen := map[uuid.UUID]bool{}
	for _, d := range g.attackCompanionsLocked(src.Controller, assign, free, false) {
		if seen[d.Attacker] || !r.countsAsCompanion(src, findBattlefieldCard(g, d.Attacker)) {
			continue
		}
		seen[d.Attacker] = true
		if have+len(seen) >= r.need {
			return true
		}
	}
	return false
}

// notAloneDeadEndLocked is the verbs' half: whether the declaration
// `after`, made from `before`, has just added l's source although its
// companion rule could never be met — not enough creatures that count
// could ever be declared beside it. A tax the player could pay counts as
// a way to add one, since the player may choose to pay it; the
// checkpoint is stricter.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) notAloneDeadEndLocked(l attackLimitSource, before, after map[uuid.UUID]uuid.UUID) bool {
	if l.source == nil || !l.sourceAttacksIn(after) || l.sourceAttacksIn(before) {
		return false
	}
	r := companionRule{need: l.limit.companionsNeeded(), greater: l.limit.OthersPowerGreater, colors: l.limit.OthersColors}
	return !g.companionsReachableLocked(l.source, r, after, false)
}

// attackCompanionsLocked lists the (creature, target) pairs some other
// creature `ap` controls could be declared attacking beside the attacks
// in `assign`: eligible (untapped, no summoning sickness, no "can't
// attack"), not already in `assign`, not one that can only attack
// alone, and with a target it may attack that no count limit refuses.
// With `free`, only a target at no CR 508.1h price counts (CR 508.1d:
// a player is never required to pay a cost to attack). In battlefield
// order; with `first`, it stops at the first one.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackCompanionsLocked(ap uuid.UUID, assign map[uuid.UUID]uuid.UUID, free, first bool) []AttackDeclaration {
	if ap == uuid.Nil || g.Battlefield == nil {
		return nil
	}
	limits := g.activeAttackLimitsLocked()
	loners := g.onlyAloneAttackersLocked()
	current := g.currentAttackAssignmentLocked()
	var targets []AttackTargetRef
	var out []AttackDeclaration
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if _, in := assign[c.InstanceID]; in || loners[c.InstanceID] {
			continue
		}
		if !AttackerEligible(c, ap) || g.attackDeclinedLocked(c.InstanceID) {
			continue
		}
		if targets == nil {
			targets = g.AttackTargetsForEffect(ap)
		}
		for _, t := range targets {
			if g.canAttackTargetWithLocked(c, t.ID) != nil {
				continue
			}
			decl := []AttackDeclaration{{Attacker: c.InstanceID, Target: t.ID}}
			if free && !g.priceAttackDeclarationLocked(decl).IsFree() {
				continue
			}
			if !countLimitsAllow(limits, current, assign, c.InstanceID, t.ID) {
				continue
			}
			out = append(out, decl[0])
			if first {
				return out
			}
		}
	}
	return out
}

// countLimitsAllow reports whether adding `id` attacking `target` to
// `assign` keeps every count limit within attackLimitRefusalLocked's
// rule, judged against the attacks there are now (`current`).
func countLimitsAllow(limits []attackLimitSource, current, assign map[uuid.UUID]uuid.UUID, id, target uuid.UUID) bool {
	if len(limits) == 0 {
		return true
	}
	after := make(map[uuid.UUID]uuid.UUID, len(assign)+1)
	for a, t := range assign {
		after[a] = t
	}
	after[id] = target
	for _, l := range limits {
		if l.limit.Scope == AttackLimitNotAlone {
			continue
		}
		n := l.countIn(after)
		if n > l.limit.Max && n > l.countIn(current) {
			return false
		}
	}
	return true
}

// attackAloneUnmetLocked is the checkpoint's question for this rule:
// the refusal when a creature in the active player's declaration has a
// companion rule the declaration does not meet (one creature that
// can't attack alone; Orcish Conscripts with one other attacker) and
// enough free companions could still be declared to meet it, or nil.
// The second result is that creature whenever there is one, refused or
// not, so the checkpoint can take it out when its rule can't be met.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackAloneUnmetLocked() (*AttackLimitError, uuid.UUID) {
	if g.attackRequirementsJudgedLocked() {
		return nil, uuid.Nil
	}
	notAlone := g.notAloneSourcesLocked()
	if len(notAlone) == 0 {
		return nil, uuid.Nil
	}
	assign := g.currentAttackAssignmentLocked()
	c, r := g.unmetNotAloneAttackerLocked(assign, notAlone)
	if c == nil {
		return nil, uuid.Nil
	}
	if !g.companionsReachableLocked(c, r, assign, true) {
		return nil, c.InstanceID
	}
	e := &AttackLimitError{Scope: AttackLimitNotAlone, Max: 1, Source: c.InstanceID, Attacker: c.InstanceID}
	e.SourceName = c.Effective().Name
	e.AttackerName = e.SourceName
	if r.need > 1 {
		e.MinOthers = r.need
	}
	e.OthersPowerGreater = r.greater
	e.OthersColors = r.colors
	return e, c.InstanceID
}

// AttackAloneUnmetForEffect is attackAloneUnmetLocked's refusal for the
// legal-move enumerator, which withholds the active player's pass while
// it is non-nil — the same function the checkpoint asks (#544).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) AttackAloneUnmetForEffect() *AttackLimitError {
	e, _ := g.attackAloneUnmetLocked()
	return e
}

// attackAloneCheckpointLocked is the checkpoint's half: refuses a
// declaration that is one creature that can't attack alone while a
// free companion could join it; when none can, takes that creature out
// of the declaration and untaps it — the declaration it was in is
// illegal, and its tap was that declaration's (CR 508.1f) — so it does
// not attack alone and the table moves on. That can happen after the
// declaration was locked in by a priority boundary inside the step (a
// trick cast, the companion killed in response), which is why it is
// not limited to a creature that has not been announced.
//
// Caller must hold g.mu in write mode with fresh layers.
func (g *Game) attackAloneCheckpointLocked() error {
	// Taking one creature out can leave another short of company (two
	// Orcish Conscripts and one more attacker), so this repeats until
	// every creature left meets its rule or one is refused. Each pass
	// removes an attacker, so it ends.
	for {
		e, lone := g.attackAloneUnmetLocked()
		if e != nil {
			return e
		}
		if lone == uuid.Nil {
			return nil
		}
		c := findBattlefieldCard(g, lone)
		if c == nil {
			return nil
		}
		g.removeFromCombatLocked(c)
		if !HasKeyword(c, "vigilance") {
			c.Tapped = false
		}
	}
}

// attackAloneCompanionsLocked is the owed answer while
// attackAloneUnmetLocked refuses the pass: every free (creature, target)
// the verb would accept beside the lone creature, keyed like
// MustAttackForEffect.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackAloneCompanionsLocked() map[uuid.UUID][]uuid.UUID {
	e, lone := g.attackAloneUnmetLocked()
	if e == nil {
		return nil
	}
	c := findBattlefieldCard(g, lone)
	if c == nil {
		return nil
	}
	r := g.notAloneSourcesLocked()[lone]
	var out map[uuid.UUID][]uuid.UUID
	for _, d := range g.attackCompanionsLocked(c.Controller, g.currentAttackAssignmentLocked(), true, false) {
		if !r.countsAsCompanion(c, findBattlefieldCard(g, d.Attacker)) {
			continue
		}
		decl := []AttackDeclaration{d}
		if g.attackLimitRefusalLocked(decl) != nil || g.attackRequirementRefusalLocked(decl) != nil {
			continue
		}
		if out == nil {
			out = map[uuid.UUID][]uuid.UUID{}
		}
		out[d.Attacker] = append(out[d.Attacker], d.Target)
	}
	return out
}

// reachWithoutLoneNotAloneLocked corrects a CR 508.1d reach whose best
// completion would leave a creature short of the company its rule needs
// (alone, for "can't attack alone"): that completion is legal only if
// enough creatures could be declared beside it for free; otherwise the
// best addition is the best one without that creature. `best`
// recomputes the addition from a candidate list.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) reachWithoutLoneNotAloneLocked(assign map[uuid.UUID]uuid.UUID, out attackRequirementReach, best func([]reqCandidate) (int, map[uuid.UUID]uuid.UUID)) attackRequirementReach {
	notAlone := g.notAloneSourcesLocked()
	if len(notAlone) == 0 || len(assign) > 0 {
		// A declaration that already has an attacker is the
		// checkpoint's to judge.
		return out
	}
	// A completion whose witness leaves a creature short of company
	// counts only if enough free companions could join it; otherwise the
	// best addition is the best one without that creature. Each pass
	// drops a candidate, so it ends.
	cands := out.cands
	for {
		c, r := g.unmetNotAloneAttackerLocked(out.witness, notAlone)
		if c == nil || g.companionsReachableLocked(c, r, out.witness, true) {
			return out
		}
		var rest []reqCandidate
		for _, cand := range cands {
			if cand.id != c.InstanceID {
				rest = append(rest, cand)
			}
		}
		if len(rest) == len(cands) {
			return out
		}
		cands = rest
		out.addition, out.witness = best(rest)
	}
}
