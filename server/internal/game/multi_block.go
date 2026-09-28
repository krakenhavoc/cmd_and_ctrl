package game

import "github.com/google/uuid"

// multi_block.go is a creature that blocks more than one attacker
// (#1706, ADR 0045 amendment of 2026-09-28, Decisions 60-63).
//
// # The rule
//
// CR 509.1a: the defending player chooses which creatures block and
// which attacker each blocks. A creature blocks one attacker, unless an
// effect says it can block more: "can block an additional creature each
// combat" (High Ground, Two-Headed Giant of Foriys), "an additional
// ninety-nine creatures" (a monstrous Hundred-Handed One) or "any
// number of creatures" (Palace Guard). CR 509.1b treats the number as
// part of the declaration's legality, like menace's minimum.
//
// CR 510.1d: a blocking creature assigns its combat damage to the
// creatures it blocks. Blocking exactly one, all of it goes there;
// blocking two or more, its controller divides it among them as they
// choose — no ordering, no lethal-first rule, no trample.
//
// # The shape
//
//   - Card.BlockingTarget stays the FIRST attacker a creature blocks, so
//     every "is it blocking" reader is untouched, and Card.AlsoBlocking
//     holds the rest. BlockedAttackers is the one reader of the set.
//   - The capacity is on the effective characteristic
//     (Characteristic.AdditionalBlocks / BlocksAnyNumber), written by
//     layer statics. BlockCapacity is the one reader.
//   - The declaration's working copy is a blockAssignment, blocker ->
//     the attackers it blocks, which the validator, the whole-combat
//     limits and the CR 509.1c requirement search all read.
//   - An entry naming a blocker that is already blocking ADDS the
//     attacker when the blocker has room, and RE-POINTS it when its
//     capacity is one (the sandbox's "re-declare blocker", unchanged).
//     A blocker with room for more and none left is refused
//     (blocker_capacity).

// BlockCapacity is how many attackers `c` can block this combat
// (CR 509.1a/b): 1 for an ordinary creature, 1 + AdditionalBlocks when
// effects let it block additional creatures, and 0 — meaning "any
// number" — when one lets it block any number. Read off the layer
// cache: a card the layer pass has not reached blocks one.
func BlockCapacity(c *Card) int {
	if c == nil || c.effective == nil {
		return 1
	}
	if c.effective.BlocksAnyNumber {
		return 0
	}
	return 1 + c.effective.AdditionalBlocks
}

// blockerHasRoom reports whether `c` could block one more attacker on
// top of the ones it already blocks.
func blockerHasRoom(c *Card) bool {
	if c == nil {
		return false
	}
	capacity := BlockCapacity(c)
	return capacity == 0 || len(c.BlockedAttackers()) < capacity
}

// BlockedAttackers is every attacker `c` blocks, in declaration order:
// BlockingTarget first, then AlsoBlocking. Nil when it is not blocking.
func (c *Card) BlockedAttackers() []uuid.UUID {
	if c == nil || c.BlockingTarget == uuid.Nil {
		return nil
	}
	out := make([]uuid.UUID, 0, 1+len(c.AlsoBlocking))
	out = append(out, c.BlockingTarget)
	return append(out, c.AlsoBlocking...)
}

// IsBlockingAttacker reports whether `c` blocks the attacker `id`.
func (c *Card) IsBlockingAttacker(id uuid.UUID) bool {
	if c == nil || id == uuid.Nil {
		return false
	}
	if c.BlockingTarget == id {
		return true
	}
	for _, a := range c.AlsoBlocking {
		if a == id {
			return true
		}
	}
	return false
}

// setBlockingSet writes the attackers `c` blocks: the first into
// BlockingTarget and the rest into AlsoBlocking. An empty set clears
// both.
func (c *Card) setBlockingSet(ids []uuid.UUID) {
	if len(ids) == 0 {
		c.clearBlocking()
		return
	}
	c.BlockingTarget = ids[0]
	c.AlsoBlocking = nil
	if len(ids) > 1 {
		c.AlsoBlocking = append([]uuid.UUID(nil), ids[1:]...)
	}
}

// clearBlocking takes `c` out of every block it is in. Every place
// that used to write BlockingTarget = uuid.Nil calls this, so the two
// fields can never disagree.
func (c *Card) clearBlocking() {
	c.BlockingTarget = uuid.Nil
	c.AlsoBlocking = nil
}

// blockAssignment is a block declaration as the validator sees it:
// blocker instance ID -> the attackers it blocks, in declaration order.
// A blocker that is not blocking has no key.
type blockAssignment map[uuid.UUID][]uuid.UUID

// has reports whether `blocker` blocks `attacker` in a.
func (a blockAssignment) has(blocker, attacker uuid.UUID) bool {
	for _, id := range a[blocker] {
		if id == attacker {
			return true
		}
	}
	return false
}

// clone is a deep copy, so a hypothetical declaration never writes
// through into the one it started from.
func (a blockAssignment) clone() blockAssignment {
	out := make(blockAssignment, len(a))
	for b, atks := range a {
		out[b] = append([]uuid.UUID(nil), atks...)
	}
	return out
}

// blockerCounts is how many creatures block each attacker in a.
func (a blockAssignment) blockerCounts() map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, atks := range a {
		for _, id := range atks {
			out[id]++
		}
	}
	return out
}

// withBlocks is `assign` with `add` staged on top, as a fresh
// assignment. Every entry ADDS its attacker to the blocker's set: the
// callers stage the requirement search's additions, which only ever
// name a blocker with room (the search offers no re-point).
func withBlocks(assign blockAssignment, add []BlockDeclaration) blockAssignment {
	out := assign.clone()
	for _, d := range add {
		if !out.has(d.Blocker, d.Attacker) {
			out[d.Blocker] = append(out[d.Blocker], d.Attacker)
		}
	}
	return out
}

// BlockingCountForEffect is how many attackers the permanent `id`
// blocks right now — Guardian of the Gateless's "for each creature
// it's blocking". 0 for a permanent that is not blocking. Attackers
// that have left the battlefield are not counted: a creature removed
// from combat is no longer being blocked by anything (CR 506.4).
//
// Caller must hold g.mu (read or write).
func (g *Game) BlockingCountForEffect(id uuid.UUID) int {
	c := findBattlefieldCard(g, id)
	if c == nil {
		return 0
	}
	n := 0
	for _, a := range c.BlockedAttackers() {
		if atk := findBattlefieldCard(g, a); atk != nil && atk.AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n
}

// liveBlockedAttackersLocked is the attackers blocker `blk` still
// blocks for damage purposes: the ones on the battlefield and still
// attacking. CR 510.1d: "If it isn't blocking any creatures (if, for
// example, they were removed from combat), it assigns no combat
// damage."
//
// Caller must hold g.mu.
func (g *Game) liveBlockedAttackersLocked(blk *Card) []uuid.UUID {
	var out []uuid.UUID
	for _, a := range blk.BlockedAttackers() {
		if atk := findBattlefieldCard(g, a); atk != nil && atk.AttackingTarget != uuid.Nil {
			out = append(out, a)
		}
	}
	return out
}

// queueBlockerDamageDivisionLocked queues the CR 510.1d prompt for a
// blocker that blocks two or more attackers: its controller divides
// its combat damage among them as they choose. It reuses the CR 510.1c
// damage-assignment prompt with the roles turned round — the frame's
// "attacker" is the blocker dealing the damage and its "blockers" are
// the attackers receiving it — and BlockerDivides set, which is what
// switches off the attacker-only rules (the lethal-first order, and
// trample). Every source keyword is snapshotted exactly as the
// attacker's prompt snapshots them, for the same died-before-resume
// reason.
//
// Caller must hold g.mu.
func (g *Game) queueBlockerDamageDivisionLocked(blk *Card, attackerIDs []uuid.UUID, power int, step string) {
	frame := &DamageAssignmentFrame{
		AttackerID:        blk.InstanceID,
		BlockerIDs:        append([]uuid.UUID(nil), attackerIDs...),
		AttackerPower:     power,
		BlockerDivides:    true,
		HasDeathtouch:     HasKeyword(blk, "deathtouch"),
		FirstStrike:       step == CombatStepFirstStrike,
		CombatStep:        step,
		SourceLifelink:    HasKeyword(blk, "lifelink"),
		SourceController:  blk.Controller,
		SourceIsCommander: blk.IsCommander,
		SourceLKI:         SourceCharacteristics(blk),
	}
	res := SourceDamageResultTraits(blk)
	frame.SourceInfect, frame.SourceWither, frame.SourceToxic = res.Infect, res.Wither, res.ToxicTotal
	g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceDamageAssignment,
		Chooser:          blk.Controller,
		Count:            len(attackerIDs),
		Source:           blk.InstanceID,
		Reason:           "Divide combat damage among the creatures it blocks",
		DamageAssignment: frame,
	})
}

// copyUUIDSlice is a fresh copy of ids, nil for an empty one.
func copyUUIDSlice(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	return append([]uuid.UUID(nil), ids...)
}

// splitAnnouncedBlocks is Game.announcedBlocks as the snapshot writes
// it (#1706): blocker -> the FIRST attacker its EventBlock named, under
// the key every binary since #830 reads, and blocker -> the rest for a
// blocker that was announced against more than one.
func splitAnnouncedBlocks(in map[uuid.UUID][]uuid.UUID) (first map[uuid.UUID]uuid.UUID, rest map[uuid.UUID][]uuid.UUID) {
	for b, atks := range in {
		if len(atks) == 0 {
			continue
		}
		if first == nil {
			first = map[uuid.UUID]uuid.UUID{}
		}
		first[b] = atks[0]
		if len(atks) > 1 {
			if rest == nil {
				rest = map[uuid.UUID][]uuid.UUID{}
			}
			rest[b] = append([]uuid.UUID(nil), atks[1:]...)
		}
	}
	return first, rest
}

// joinAnnouncedBlocks is splitAnnouncedBlocks' inverse. A file written
// before #1706 has no `rest`, and restores each blocker with its one
// announced pair.
func joinAnnouncedBlocks(first map[uuid.UUID]uuid.UUID, rest map[uuid.UUID][]uuid.UUID) map[uuid.UUID][]uuid.UUID {
	if len(first) == 0 {
		return nil
	}
	out := make(map[uuid.UUID][]uuid.UUID, len(first))
	for b, a := range first {
		out[b] = append([]uuid.UUID{a}, rest[b]...)
	}
	return out
}
