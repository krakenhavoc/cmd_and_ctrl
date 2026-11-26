package game

import (
	"slices"

	"github.com/google/uuid"
)

// combat_reassign.go — #6193: blocks changed after blockers are declared.
//
//   - Balduvian Warlord: "Remove target blocking creature from combat.
//     Creatures it was blocking that hadn't become blocked by another
//     creature this combat become unblocked, then it blocks an attacking
//     creature of your choice." (Ydwen Efreet and False Orders say "had
//     become blocked by only that creature this combat": the same test.)
//   - Sorrow's Path: two blockers that could each block all the creatures
//     the other is blocking swap what they block.
//   - General Jarkeld: two blocked attackers that could each be blocked by
//     all the creatures blocking the other swap their blockers.
//
// THE RECORD. CR 509.1h: an attacker stays blocked "until the creature is
// removed from combat, an effect says that it becomes blocked or
// unblocked, or the combat phase ends", and "remains blocked even if all
// the creatures blocking it are removed from combat". Becoming unblocked is
// therefore an effect's word, and it is the blocked record
// (Game.blockedAttackers) losing the attacker: BecomeUnblockedForEffect.
// Whether an attacker "had become blocked by another creature this combat"
// is history the live blocking fields lose once that creature leaves
// combat, so every block is noted as it is made (TurnTally.CombatBlocks,
// with the combat phase it was made in).
//
// BLOCKING BY AN EFFECT. A creature an effect makes block is a blocking
// creature without a declaration (CR 509.3a-d), so no CR 509.1 requirement
// or count is asked of it. The issue's plan reuses the pair legality the
// declaration uses (BlockPairRefusalLocked: restrictions, evasion,
// protection), and the CR 506.3e / 509.4a rule that a creature can block
// only an attacker attacking its controller, a planeswalker that player
// controls or a battle that player protects. The triggers follow CR
// 509.3: one EventBlock per attacker it was not already blocking (Amount
// is its place in the creature's set, so "whenever ~ blocks" fires only
// for a creature that was not blocking at that time, CR 509.3a), and one
// EventBecomesBlocked for an attacker that was an unblocked creature (CR
// 509.3c), as the declaration's lock-in emits them. An attacker that was
// already blocked does not become blocked again (#830).

// CombatBlockRecord is one block made this turn: the blocking object, the
// attacking object, and the combat phase (Turn.PhaseID) it was made in.
type CombatBlockRecord struct {
	Blocker  ObjectRef `json:"blocker"`
	Attacker ObjectRef `json:"attacker"`
	PhaseID  int       `json:"phaseId,omitempty"`
}

// noteCombatBlockLocked records that `blocker` blocks `attacker` in this
// combat. A fresh slice, never an append into shared capacity, for the
// undo clone's sake (#1238). Caller must hold g.mu in write mode.
func (g *Game) noteCombatBlockLocked(blocker, attacker uuid.UUID) {
	if blocker == uuid.Nil || attacker == uuid.Nil {
		return
	}
	rec := CombatBlockRecord{
		Blocker:  ObjectRef{ID: blocker, Epoch: g.objectEpochLocked(blocker)},
		Attacker: ObjectRef{ID: attacker, Epoch: g.objectEpochLocked(attacker)},
		PhaseID:  g.Turn.PhaseID,
	}
	if slices.Contains(g.TurnTally.CombatBlocks, rec) {
		return
	}
	out := make([]CombatBlockRecord, len(g.TurnTally.CombatBlocks), len(g.TurnTally.CombatBlocks)+1)
	copy(out, g.TurnTally.CombatBlocks)
	g.TurnTally.CombatBlocks = append(out, rec)
}

// blockedThisCombatByOtherThanLocked reports whether the attacker `atk`,
// as the object it is now, was blocked this combat by a creature other
// than `blocker`. Caller must hold g.mu.
func (g *Game) blockedThisCombatByOtherThanLocked(atk, blocker uuid.UUID) bool {
	epoch := g.objectEpochLocked(atk)
	for _, r := range g.TurnTally.CombatBlocks {
		if r.PhaseID == g.Turn.PhaseID && r.Attacker.ID == atk && r.Attacker.Epoch == epoch && r.Blocker.ID != blocker {
			return true
		}
	}
	return false
}

// BecomeUnblockedForEffect makes the attacking creature `attacker` an
// UNBLOCKED creature (CR 509.1h: "an effect says that it becomes ...
// unblocked"): it leaves the blocked record, so it deals its combat damage
// to what it attacks. Refused (false) outside the declare-blockers step
// and the combat steps after it, for a permanent that is not attacking or
// not blocked, and while a creature is still blocking it, which no printed
// effect asks for. Caller must hold g.mu in write mode.
func (g *Game) BecomeUnblockedForEffect(attacker uuid.UUID) bool {
	if !g.blockersDeclaredStepLocked() || !g.attackerBlockedLocked(attacker) {
		return false
	}
	a := findBattlefieldCard(g, attacker)
	if a == nil || a.AttackingTarget == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].IsBlockingAttacker(attacker) {
			return false
		}
	}
	delete(g.blockedAttackers, attacker)
	g.invalidateLayersForAttackChangeLocked()
	return true
}

// RemoveBlockerUnblockingForEffect is "remove target blocking creature
// from combat; creatures it was blocking that hadn't become blocked by
// another creature this combat become unblocked" (Balduvian Warlord,
// Ydwen Efreet, False Orders). Returns the attackers that became
// unblocked. A creature that is not blocking is still removed from combat
// if it is attacking, and unblocks nothing.
//
// Caller must hold g.mu in write mode.
func (g *Game) RemoveBlockerUnblockingForEffect(blocker uuid.UUID) []uuid.UUID {
	c := findBattlefieldCard(g, blocker)
	if c == nil {
		return nil
	}
	atks := c.BlockedAttackers()
	if c.AttackingTarget == uuid.Nil && len(atks) == 0 {
		return nil
	}
	g.removeFromCombatLocked(c)
	var out []uuid.UUID
	for _, a := range atks {
		if g.blockedThisCombatByOtherThanLocked(a, blocker) {
			continue
		}
		if g.BecomeUnblockedForEffect(a) {
			out = append(out, a)
		}
	}
	return out
}

// CanBlockByEffectForEffect reports whether an effect may make `blocker`
// block `attacker` now: both on the battlefield, the blocker a creature,
// the attacker attacking its controller (or that player's planeswalker or
// battle, CR 506.3e), and the pair legal under the declaration's pair
// rules (BlockPairRefusalLocked). Sorrow's Path's and General Jarkeld's
// "could block" is this question. Caller must hold g.mu with fresh layers.
func (g *Game) CanBlockByEffectForEffect(blocker, attacker uuid.UUID) bool {
	b, a := findBattlefieldCard(g, blocker), findBattlefieldCard(g, attacker)
	if b == nil || a == nil || !b.IsCreature() || a.AttackingTarget == uuid.Nil {
		return false
	}
	if g.defendingPlayerForAttackerLocked(a) != b.Controller {
		return false
	}
	return g.BlockPairRefusalLocked(a, b).Legal()
}

// BlockAttackersForEffect makes `blocker` block each of `attackers` it is
// not already blocking and may block (CanBlockByEffectForEffect), as an
// effect rather than a declaration: Balduvian Warlord's "then it blocks an
// attacking creature of your choice", False Orders' "you may have it block
// an attacking creature of your choice", Sorrow's Path's swap. Returns how
// many attackers it now blocks that it did not before. Nothing happens
// outside the declare-blockers step and the combat steps after it.
//
// Caller must hold g.mu in write mode.
func (g *Game) BlockAttackersForEffect(blocker uuid.UUID, attackers []uuid.UUID) int {
	if !g.blockersDeclaredStepLocked() {
		return 0
	}
	g.RecomputeLayersIfStaleLocked()
	c := findBattlefieldCard(g, blocker)
	if c == nil {
		return 0
	}
	set := c.BlockedAttackers()
	type pair struct {
		attacker uuid.UUID
		nth      int
		becomes  bool
	}
	var added []pair
	for _, a := range g.withBandMatesLocked(attackers) {
		if a == uuid.Nil || slices.Contains(set, a) || !g.CanBlockByEffectForEffect(blocker, a) {
			continue
		}
		// Asked BEFORE the change: once this creature blocks it, the
		// attacker no longer reads as unblocked (CR 509.3c).
		becomes := g.UnblockedAttackerForEffect(a)
		set = append(set, a)
		added = append(added, pair{attacker: a, nth: len(set), becomes: becomes})
	}
	if len(added) == 0 {
		return 0
	}
	c = findBattlefieldCard(g, blocker)
	c.setBlockingSet(set)
	if g.announcedBlocks == nil {
		g.announcedBlocks = map[uuid.UUID][]uuid.UUID{}
	}
	g.announcedBlocks[blocker] = append([]uuid.UUID(nil), set...)
	if g.blockedAttackers == nil {
		g.blockedAttackers = map[uuid.UUID]bool{}
	}
	actor := c.Controller
	for _, p := range added {
		g.blockedAttackers[p.attacker] = true
		g.noteCombatBlockLocked(blocker, p.attacker)
	}
	g.invalidateLayersForAttackChangeLocked()
	for _, p := range added {
		g.EmitEvent(Event{Kind: EventBlock, Actor: actor, Source: blocker, CardID: blocker, Target: p.attacker, Amount: p.nth})
	}
	for _, p := range added {
		if p.becomes {
			g.EmitEvent(Event{Kind: EventBecomesBlocked, Actor: actor, Source: p.attacker, CardID: p.attacker, Target: p.attacker})
		}
	}
	return len(added)
}

// SwapBlockersForEffect is Sorrow's Path: "If each of those creatures
// could block all creatures that the other is blocking, remove both of
// them from combat. Each one then blocks all creatures the other was
// blocking." Returns false, changing nothing, when either is not a
// blocking creature or one could not block every creature the other
// blocks. The attackers stay blocked throughout (CR 509.1h), so none of
// them "becomes blocked" again.
//
// Caller must hold g.mu in write mode.
func (g *Game) SwapBlockersForEffect(first, second uuid.UUID) bool {
	if !g.blockersDeclaredStepLocked() || first == second {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	a, b := findBattlefieldCard(g, first), findBattlefieldCard(g, second)
	if a == nil || b == nil {
		return false
	}
	aSet, bSet := a.BlockedAttackers(), b.BlockedAttackers()
	if len(aSet) == 0 || len(bSet) == 0 {
		return false
	}
	for _, atk := range bSet {
		if !g.CanBlockByEffectForEffect(first, atk) {
			return false
		}
	}
	for _, atk := range aSet {
		if !g.CanBlockByEffectForEffect(second, atk) {
			return false
		}
	}
	g.removeFromCombatLocked(findBattlefieldCard(g, first))
	g.removeFromCombatLocked(findBattlefieldCard(g, second))
	g.BlockAttackersForEffect(first, bSet)
	g.BlockAttackersForEffect(second, aSet)
	return true
}

// SwitchBlockedAttackersForEffect is General Jarkeld: "If each of those
// creatures could be blocked by all creatures that the other is blocked
// by, each creature that's blocking exactly one of those attacking
// creatures stops blocking it and is blocking the other attacking
// creature." Returns false, changing nothing, when either is not a blocked
// attacking creature or a blocker of one could not block the other. A
// creature blocking both is left as it is. Each switched blocker gets one
// EventBlock for the attacker it now blocks (CR 509.3b, 509.3d); it was a
// blocking creature throughout, so "whenever ~ blocks" does not fire (CR
// 509.3a), and both attackers were blocked throughout.
//
// Caller must hold g.mu in write mode.
func (g *Game) SwitchBlockedAttackersForEffect(x, y uuid.UUID) bool {
	if !g.blockersDeclaredStepLocked() || x == y || !g.attackerBlockedLocked(x) || !g.attackerBlockedLocked(y) {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	var ofX, ofY []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		bx, by := c.IsBlockingAttacker(x), c.IsBlockingAttacker(y)
		if bx {
			ofX = append(ofX, c.InstanceID)
		}
		if by {
			ofY = append(ofY, c.InstanceID)
		}
	}
	for _, b := range ofX {
		if !g.CanBlockByEffectForEffect(b, y) {
			return false
		}
	}
	for _, b := range ofY {
		if !g.CanBlockByEffectForEffect(b, x) {
			return false
		}
	}
	type moved struct {
		blocker, to uuid.UUID
		nth         int
		actor       uuid.UUID
	}
	var events []moved
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		bx, by := c.IsBlockingAttacker(x), c.IsBlockingAttacker(y)
		if bx == by {
			continue
		}
		from, to := x, y
		if by {
			from, to = y, x
		}
		set := c.BlockedAttackers()
		for j, a := range set {
			if a == from {
				set[j] = to
				events = append(events, moved{blocker: c.InstanceID, to: to, nth: j + 1, actor: c.Controller})
			}
		}
		c.setBlockingSet(set)
		if g.announcedBlocks == nil {
			g.announcedBlocks = map[uuid.UUID][]uuid.UUID{}
		}
		g.announcedBlocks[c.InstanceID] = append([]uuid.UUID(nil), set...)
	}
	for _, m := range events {
		g.noteCombatBlockLocked(m.blocker, m.to)
	}
	g.invalidateLayersForAttackChangeLocked()
	for _, m := range events {
		// nth of at least 2 would also keep a "whenever ~ blocks" from
		// firing; a creature that was blocking throughout is not a new
		// blocker even in its first slot (CR 509.3a), so Amount is
		// pinned past the first.
		n := m.nth
		if n < 2 {
			n = 2
		}
		g.EmitEvent(Event{Kind: EventBlock, Actor: m.actor, Source: m.blocker, CardID: m.blocker, Target: m.to, Amount: n})
	}
	return true
}
