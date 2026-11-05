package game

import "github.com/google/uuid"

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

// notAloneSourcesLocked is every creature an active "can't attack
// alone" limit is about, or nil — the answer at nearly every table.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) notAloneSourcesLocked() map[uuid.UUID]bool {
	var out map[uuid.UUID]bool
	for _, l := range g.activeAttackLimitsLocked() {
		if l.limit.Scope != AttackLimitNotAlone || l.source == nil {
			continue
		}
		if out == nil {
			out = map[uuid.UUID]bool{}
		}
		out[l.source.InstanceID] = true
	}
	return out
}

// loneNotAloneAttacker reports the one attacker in `assign` when it is
// the only one and it can't attack alone, or uuid.Nil.
func loneNotAloneAttacker(assign map[uuid.UUID]uuid.UUID, notAlone map[uuid.UUID]bool) uuid.UUID {
	if len(assign) != 1 {
		return uuid.Nil
	}
	for id := range assign {
		if notAlone[id] {
			return id
		}
	}
	return uuid.Nil
}

// notAloneDeadEndLocked is the verbs' half: whether the declaration
// `after`, made from `before`, has begun with l's source as its only
// attacker although no other creature could ever be added beside it.
// A tax the player could pay counts as a way to add one, since the
// player may choose to pay it; the checkpoint is stricter.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) notAloneDeadEndLocked(l attackLimitSource, before, after map[uuid.UUID]uuid.UUID) bool {
	if l.source == nil || len(after) != 1 || !l.sourceAttacksIn(after) || l.sourceAttacksIn(before) {
		return false
	}
	return !g.attackCompanionAvailableLocked(l.source.Controller, after, false)
}

// attackCompanionAvailableLocked reports whether some other creature
// `ap` controls could be declared attacking beside the attacks in
// `assign`: eligible (untapped, no summoning sickness, no "can't
// attack"), not already in `assign`, not one that can only attack
// alone, and with a target it may attack that no count limit refuses.
// With `free`, only a target at no CR 508.1h price counts (CR 508.1d:
// a player is never required to pay a cost to attack).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackCompanionAvailableLocked(ap uuid.UUID, assign map[uuid.UUID]uuid.UUID, free bool) bool {
	return len(g.attackCompanionsLocked(ap, assign, free, true)) > 0
}

// attackCompanionsLocked lists the (creature, target) pairs
// attackCompanionAvailableLocked accepts, in battlefield order; with
// `first`, it stops at the first one.
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
// the refusal when the active player's declaration is one creature that
// can't attack alone and a free companion could still be declared, or
// nil. The second result is that lone creature whenever there is one,
// refused or not, so the checkpoint can take it out when nothing can
// join it.
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
	lone := loneNotAloneAttacker(assign, notAlone)
	if lone == uuid.Nil {
		return nil, uuid.Nil
	}
	c := findBattlefieldCard(g, lone)
	if c == nil || !g.attackCompanionAvailableLocked(c.Controller, assign, true) {
		return nil, lone
	}
	e := &AttackLimitError{Scope: AttackLimitNotAlone, Max: 1, Source: lone, Attacker: lone}
	e.SourceName = c.Effective().Name
	e.AttackerName = e.SourceName
	return e, lone
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
	return nil
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
	var out map[uuid.UUID][]uuid.UUID
	for _, d := range g.attackCompanionsLocked(c.Controller, g.currentAttackAssignmentLocked(), true, false) {
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
// completion would leave a creature that can't attack alone as the only
// attacker: that completion is legal only if some other creature can be
// declared beside it for free; otherwise the best addition is the best
// one without any such creature. `best` recomputes the addition from a
// candidate list.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) reachWithoutLoneNotAloneLocked(ap uuid.UUID, assign map[uuid.UUID]uuid.UUID, out attackRequirementReach, best func([]reqCandidate) (int, map[uuid.UUID]uuid.UUID)) attackRequirementReach {
	notAlone := g.notAloneSourcesLocked()
	if len(notAlone) == 0 || len(assign) > 0 {
		// A declaration that already has an attacker: either the lone
		// creature is in it, which the checkpoint judges, or any
		// addition makes two.
		return out
	}
	if loneNotAloneAttacker(out.witness, notAlone) == uuid.Nil {
		return out
	}
	if g.attackCompanionAvailableLocked(ap, out.witness, true) {
		return out
	}
	var rest []reqCandidate
	for _, c := range out.cands {
		if !notAlone[c.id] {
			rest = append(rest, c)
		}
	}
	out.addition, out.witness = best(rest)
	return out
}
