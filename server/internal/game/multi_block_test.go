package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// multi_block_test.go pins #1706 (ADR 0045 amendment of 2026-09-28,
// Decisions 60-63): a creature that can block more than one attacker.
// The capacity is read off the effective characteristic, the verb adds
// attackers up to it and refuses past it, the count bounds, limits and
// CR 509.1c requirements see a blocker's whole set, the option
// generator offers only legal sets, the blocker divides its damage
// (CR 510.1d), and the set survives undo and a snapshot — including one
// written before the set existed. The cards, the enumerator, the bot
// and the view are pinned in cards/effects, legal, aiseat and protocol.

const (
	blockTwoOracle = "test-multi-block-two"
	blockAnyOracle = "test-multi-block-any"
	// blockTwoMustBlockOracle is a capacity-2 creature that also
	// "blocks each combat if able".
	blockTwoMustBlockOracle = "test-multi-block-two-must"
)

// withMultiBlockStatics stubs the catalog statics so a card with one of
// the test oracle IDs can block more than one attacker — the same
// layer-6 statics effects.CanBlockAdditional / CanBlockAnyNumber build.
func withMultiBlockStatics(t *testing.T) {
	t.Helper()
	self := func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID }
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case blockTwoOracle:
			return []StaticAbility{{Layer: Layer6Ability, AppliesTo: self,
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.AdditionalBlocks++ }}}
		case blockAnyOracle:
			return []StaticAbility{{Layer: Layer6Ability, AppliesTo: self,
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.BlocksAnyNumber = true }}}
		case blockTwoMustBlockOracle:
			return []StaticAbility{{Layer: Layer6Ability, AppliesTo: self,
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.AdditionalBlocks++
					c.BlockRequirements = append(c.BlockRequirements, BlockRequirement{Kind: BlockRequirementBlocks, SourceName: "Test"})
				}}}
		}
		return nil
	})
}

// pushMultiBlocker seeds a creature carrying one of the test oracle IDs
// and fires the zone move, so the layer pass reaches it.
func pushMultiBlocker(t *testing.T, g *Game, owner *Player, name, oracleID string, power, toughness int) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.OracleID = oracleID
	c.Power, c.Toughness = power, toughness
	return pushTypedTestCard(g, c)
}

// blockedSet is the attackers `id` blocks, read under the lock.
func blockedSet(g *Game, id uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		if c := findBattlefieldCard(g, id); c != nil {
			out = c.BlockedAttackers()
		}
	})
	return out
}

func sameIDs(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// capacityRefusal asserts err is a blocker_capacity refusal of `n`.
func capacityRefusal(t *testing.T, err error, n int) *BlockRefusedError {
	t.Helper()
	var br *BlockRefusedError
	if !errors.As(err, &br) || br.Reason != BlockReasonBlockerCapacity {
		t.Fatalf("err = %v, want a blocker_capacity refusal", err)
	}
	if br.N != n {
		t.Errorf("refusal N = %d, want %d", br.N, n)
	}
	return br
}

// TestMultiBlockCapacityTwoBlocksTwoAndRefusesAThird — "can block an
// additional creature each combat": a second attacker is ADDED, and a
// third is refused with the capacity, whether it arrives on its own or
// inside one declaration of three.
func TestMultiBlockCapacityTwoBlocksTwoAndRefusesAThird(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	c := pushCombatant(t, g, me, "Bear C", 2, 2)
	giant := pushMultiBlocker(t, g, opp, "Two-Headed Test", blockTwoOracle, 4, 6)
	declareAttacks(t, g, a, b, c)

	// All three in one declaration: refused whole, nothing stored.
	capacityRefusal(t, block(t, g, giant, a, giant, b, giant, c), 2)
	if got := blockedSet(g, giant); len(got) != 0 {
		t.Fatalf("a refused declaration stored %v", got)
	}

	if err := block(t, g, giant, a); err != nil {
		t.Fatalf("first attacker: %v", err)
	}
	if err := block(t, g, giant, b); err != nil {
		t.Fatalf("second attacker: %v", err)
	}
	if got := blockedSet(g, giant); !sameIDs(got, []uuid.UUID{a, b}) {
		t.Fatalf("blocks %v, want [A B]: the second entry must ADD, not re-point", got)
	}
	br := capacityRefusal(t, block(t, g, giant, c), 2)
	if got := br.Sentence(opp.ID); got != "Two-Headed Test can't block more than two creatures each combat." {
		t.Errorf("sentence = %q", got)
	}
	if got := blockedSet(g, giant); !sameIDs(got, []uuid.UUID{a, b}) {
		t.Fatalf("a refused third block changed the set to %v", got)
	}
	// Re-declaring a pairing already stored changes nothing.
	if err := block(t, g, giant, a); err != nil {
		t.Fatalf("re-declaring a stored pairing: %v", err)
	}
	if got := blockedSet(g, giant); !sameIDs(got, []uuid.UUID{a, b}) {
		t.Fatalf("an idempotent entry changed the set to %v", got)
	}
}

// TestMultiBlockAnyNumberBlocksEveryAttacker — "can block any number of
// creatures": one declaration puts it on every attacker.
func TestMultiBlockAnyNumberBlocksEveryAttacker(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var atks []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		atks = append(atks, pushCombatant(t, g, me, "Bear "+n, 1, 1))
	}
	guard := pushMultiBlocker(t, g, opp, "Palace Test", blockAnyOracle, 1, 20)
	declareAttacks(t, g, atks...)
	var pairs []uuid.UUID
	for _, a := range atks {
		pairs = append(pairs, guard, a)
	}
	if err := block(t, g, pairs...); err != nil {
		t.Fatalf("any number: %v", err)
	}
	if got := blockedSet(g, guard); !sameIDs(got, atks) {
		t.Fatalf("blocks %v, want all five", got)
	}
}

// TestMultiBlockOrdinaryBlockerStillRePoints — a creature that blocks
// one attacker keeps the sandbox's re-point: its second entry moves it.
func TestMultiBlockOrdinaryBlockerStillRePoints(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	w := pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, a, b)
	if err := block(t, g, w, a); err != nil {
		t.Fatal(err)
	}
	if err := block(t, g, w, b); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	if got := blockedSet(g, w); !sameIDs(got, []uuid.UUID{b}) {
		t.Fatalf("blocks %v, want only B", got)
	}
}

// TestMultiBlockAnnouncesOnePairEach — the lock-in announces one
// EventBlock per attacker, numbered in Amount so "whenever this blocks"
// (CR 509.3a) can fire once, and each attacker becomes blocked once.
func TestMultiBlockAnnouncesOnePairEach(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	giant := pushMultiBlocker(t, g, opp, "Two-Headed Test", blockTwoOracle, 4, 6)
	declareAttacks(t, g, a, b)
	seq := g.Events[len(g.Events)-1].Seq
	if err := block(t, g, giant, a, giant, b); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	var blocks, became []Event
	for _, ev := range g.Events {
		if ev.Seq <= seq {
			continue
		}
		switch ev.Kind {
		case EventBlock:
			blocks = append(blocks, ev)
		case EventBecomesBlocked:
			became = append(became, ev)
		}
	}
	if len(blocks) != 2 || blocks[0].Target != a || blocks[0].Amount != 1 || blocks[1].Target != b || blocks[1].Amount != 2 {
		t.Fatalf("EventBlock = %+v, want A (1) then B (2)", blocks)
	}
	if len(became) != 2 {
		t.Fatalf("%d EventBecomesBlocked, want one per attacker", len(became))
	}
	// A second lock-in announces nothing new.
	g.WithWriteLock(func() {
		if g.blockDeclarationPendingLocked() {
			t.Error("an announced multi-block still reads as pending")
		}
	})
}

// TestMultiBlockDividesItsDamageAndTakesBoth — CR 510.1d: the blocker
// takes both attackers' damage and its controller divides its own among
// them, as they choose — a split the attacker's lethal-first order
// would refuse is accepted, and trample is not on offer.
func TestMultiBlockDividesItsDamageAndTakesBoth(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	giant := pushMultiBlocker(t, g, opp, "Two-Headed Test", blockTwoOracle, 3, 6)
	declareAttacks(t, g, a, b)
	if err := block(t, g, giant, a, giant, b); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)

	if got := damageOn(g, giant); got != 4 {
		t.Errorf("blocker has %d damage, want 4 — both attackers' 2", got)
	}
	var choice *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceDamageAssignment {
			choice = c
		}
	}
	if choice == nil {
		t.Fatal("no CR 510.1d division prompt")
	}
	f := choice.DamageAssignment
	if choice.Chooser != opp.ID || f.AttackerID != giant || !f.BlockerDivides || f.AllowTrample || f.AttackerPower != 3 || !sameIDs(f.BlockerIDs, []uuid.UUID{a, b}) {
		t.Fatalf("prompt = chooser %s, frame %+v", choice.Chooser, *f)
	}
	if got := damageOn(g, a) + damageOn(g, b); got != 0 {
		t.Fatalf("the blocker dealt %d before its controller divided it", got)
	}
	if err := g.ResolveDamageAssignment(choice.ID, opp.ID,
		[]DamageAssignmentEntry{{BlockerID: a, Amount: 3}, {BlockerID: b, Amount: 0}}, 1); err == nil {
		t.Fatal("trample from a blocker was accepted")
	}
	// 1 to A (short of lethal) then 2 to B: the lethal-first order would
	// refuse this; a blocker's division is free.
	if err := g.ResolveDamageAssignment(choice.ID, opp.ID,
		[]DamageAssignmentEntry{{BlockerID: a, Amount: 1}, {BlockerID: b, Amount: 2}}, 0); err != nil {
		t.Fatalf("a free division: %v", err)
	}
	if got := damageOn(g, a); got != 1 {
		t.Errorf("A has %d damage, want 1", got)
	}
	if findBattlefieldCard(g, b) != nil {
		t.Error("B took lethal 2 and is still on the battlefield")
	}
	if findBattlefieldCard(g, giant) == nil {
		t.Error("the blocker died to 4 damage on 6 toughness")
	}
}

// TestMultiBlockOneLiveAttackerTakesItAll — CR 510.1d: a blocker whose
// other attacker has left combat is blocking one creature, and deals it
// all its damage with no prompt.
func TestMultiBlockOneLiveAttackerTakesItAll(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 5)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	giant := pushMultiBlocker(t, g, opp, "Two-Headed Test", blockTwoOracle, 3, 6)
	declareAttacks(t, g, a, b)
	if err := block(t, g, giant, a, giant, b); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.removeFromCombatLocked(findBattlefieldCard(g, b)) })
	advanceIntoStep(t, g, StepCombatDamage)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceDamageAssignment {
			t.Fatal("a prompt for a blocker with one attacker left")
		}
	}
	if got := damageOn(g, a); got != 3 {
		t.Errorf("A has %d damage, want all 3", got)
	}
}

// TestMultiBlockMenaceStillNeedsTwoBlockers — one creature blocking
// several attackers is still ONE blocker of each: a menace attacker is
// not blocked by an any-number blocker alone.
func TestMultiBlockMenaceStillNeedsTwoBlockers(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m := pushCombatant(t, g, me, "Menace Bear", 2, 2, "menace")
	plain := pushCombatant(t, g, me, "Bear", 2, 2)
	guard := pushMultiBlocker(t, g, opp, "Palace Test", blockAnyOracle, 1, 20)
	w := pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, m, plain)
	if err := block(t, g, guard, plain); err != nil {
		t.Fatal(err)
	}
	var br *BlockRefusedError
	if err := block(t, g, guard, m); !errors.As(err, &br) || br.Reason != BlockReasonTooFewBlockers {
		t.Fatalf("a lone multi-blocker on a menace attacker: %v, want too_few_blockers", err)
	}
	if err := block(t, g, guard, m, w, m); err != nil {
		t.Fatalf("the multi-blocker and a second creature: %v", err)
	}
	if got := blockedSet(g, guard); !sameIDs(got, []uuid.UUID{plain, m}) {
		t.Fatalf("guard blocks %v", got)
	}
}

// TestMultiBlockLureIsCountedPerAttacker — CR 509.1c with a blocker
// that can take both Lure'd attackers: each Lure asks it to block, so
// blocking one is a step and the pass is refused until it blocks the
// other; the flow's witness is the second block.
func TestMultiBlockLureIsCountedPerAttacker(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Lured A", 1, 1)
	b := pushCombatant(t, g, me, "Lured B", 1, 1)
	withBlockRequirement(t, g, a, BlockRequirementLure)
	withBlockRequirement(t, g, b, BlockRequirementLure)
	guard := pushMultiBlocker(t, g, opp, "Palace Test", blockAnyOracle, 1, 20)
	declareAttacks(t, g, a, b)

	var witness []BlockDeclaration
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked(); witness = g.blockRequirementWitnessLocked(opp.ID) })
	if len(witness) != 2 {
		t.Fatalf("witness = %+v, want the guard on both", witness)
	}
	if err := block(t, g, guard, a); err != nil {
		t.Fatal(err)
	}
	passToDefender(t, g)
	br := blockRequirementErr(t, g.PassPriority())
	if br.Blocker != guard || br.Attacker != b {
		t.Errorf("the pass's refusal names %s -> %s, want the guard on B", br.Blocker, br.Attacker)
	}
	if err := block(t, g, guard, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both Lures obeyed: %v", err)
	}
}

// TestMultiBlockBlocksEachCombatCountsOnce — "blocks each combat if
// able" is ONE requirement however many attackers the creature takes:
// blocking one obeys it, and the pass is accepted with the second
// attacker left alone.
func TestMultiBlockBlocksEachCombatCountsOnce(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 1, 1)
	b := pushCombatant(t, g, me, "Bear B", 1, 1)
	giant := pushMultiBlocker(t, g, opp, "Must Block Test", blockTwoMustBlockOracle, 3, 6)
	declareAttacks(t, g, a, b)
	var obeyed int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		obeyed = len(g.blockRequirementHitsLocked(opp.ID, blockAssignment{giant: {a, b}}))
	})
	if obeyed != 1 {
		t.Errorf("a creature blocking two obeys its one requirement %d times", obeyed)
	}
	if err := block(t, g, giant, a); err != nil {
		t.Fatal(err)
	}
	passToDefender(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass after one block: %v — the requirement is met", err)
	}
}

// TestMultiBlockLimitCountsTheCreatureOnce — a whole-combat limit
// counts CREATURES that block: one blocking two attackers under "no
// more than one creature can block" is legal, and a second creature is
// refused.
func TestMultiBlockLimitCountsTheCreatureOnce(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(1)})
	pushLimitSource(t, g, me, "Silent Test")
	a := pushCombatant(t, g, me, "Bear A", 1, 1)
	b := pushCombatant(t, g, me, "Bear B", 1, 1)
	giant := pushMultiBlocker(t, g, opp, "Two-Headed Test", blockTwoOracle, 3, 6)
	w := pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, a, b)
	if err := block(t, g, giant, a, giant, b); err != nil {
		t.Fatalf("one creature blocking two under a one-creature limit: %v", err)
	}
	var br *BlockRefusedError
	if err := block(t, g, w, a); !errors.As(err, &br) || br.Reason != BlockReasonDeclarationLimit {
		t.Fatalf("a second creature: %v, want declaration_limit", err)
	}
}

// TestMultiBlockOptionsAreLegalAndNeverRepeat — the one option
// generator keeps a multi-blocker eligible while it has room, never
// offers it an attacker it already blocks, stops once it is full, and
// every option passes the verb's validator.
func TestMultiBlockOptionsAreLegalAndNeverRepeat(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	c := pushCombatant(t, g, me, "Bear C", 2, 2)
	giant := pushMultiBlocker(t, g, opp, "Two-Headed Test", blockTwoOracle, 4, 6)
	declareAttacks(t, g, a, b, c)

	optionsFor := func() map[uuid.UUID]bool {
		out := map[uuid.UUID]bool{}
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			base := g.currentBlockAssignmentLocked()
			for _, o := range g.BlockOptionsLocked(opp.ID, 12) {
				if _, err := g.checkBlockDeclarationLocked(base, o.Blocks); err != nil {
					t.Errorf("option %+v is refused: %v", o.Blocks, err)
				}
				for _, d := range o.Blocks {
					if d.Blocker == giant {
						out[d.Attacker] = true
					}
				}
			}
		})
		return out
	}
	if got := optionsFor(); len(got) != 3 {
		t.Fatalf("offered the giant %v, want all three attackers", got)
	}
	if err := block(t, g, giant, a); err != nil {
		t.Fatal(err)
	}
	if got := optionsFor(); got[a] || !got[b] || !got[c] {
		t.Fatalf("with A blocked the giant is offered %v, want B and C only", got)
	}
	if err := block(t, g, giant, b); err != nil {
		t.Fatal(err)
	}
	if got := optionsFor(); len(got) != 0 {
		t.Fatalf("a full giant is still offered %v", got)
	}
}

// TestMultiBlockSurvivesUndoAndSnapshot — the whole set rides Clone /
// RestoreFrom and the persisted snapshot, with the announcements, so a
// restored lock-in re-announces nothing.
func TestMultiBlockSurvivesUndoAndSnapshot(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	giant := pushMultiBlocker(t, g, opp, "Two-Headed Test", blockTwoOracle, 4, 6)
	declareAttacks(t, g, a, b)
	saved := g.Clone()
	if err := block(t, g, giant, a, giant, b); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	blocked := g.Clone()

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := blockedSet(restored, giant); !sameIDs(got, []uuid.UUID{a, b}) {
		t.Fatalf("snapshot restored %v", got)
	}
	restored.WithWriteLock(func() {
		if got := restored.announcedBlocks[giant]; !sameIDs(got, []uuid.UUID{a, b}) {
			t.Errorf("announced %v after the snapshot", got)
		}
		if restored.blockDeclarationPendingLocked() {
			t.Error("a restored lock-in would announce the block again")
		}
	})

	g.RestoreFrom(saved)
	if got := blockedSet(g, giant); len(got) != 0 {
		t.Fatalf("undo left %v", got)
	}
	g.RestoreFrom(blocked)
	if got := blockedSet(g, giant); !sameIDs(got, []uuid.UUID{a, b}) {
		t.Fatalf("redo restored %v", got)
	}
	// A clone does not alias the live set. (RestoreFrom ADOPTS its
	// argument, so the check takes a fresh one.)
	fresh := g.Clone()
	g.WithWriteLock(func() { findBattlefieldCard(g, giant).AlsoBlocking[0] = uuid.Nil })
	if got := blockedSet(fresh, giant); !sameIDs(got, []uuid.UUID{a, b}) {
		t.Fatalf("the clone shares its set with the live game: %v", got)
	}
}

// TestMultiBlockOldSnapshotRestoresASingleBlock — a snapshot written
// before #1706 has blockingTarget and announcedBlocks as single IDs and
// no alsoBlocking key; it restores the one block it describes.
func TestMultiBlockOldSnapshotRestoresASingleBlock(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear", 2, 2)
	w := pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, a)
	if err := block(t, g, w, a); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	// Strip what only this binary writes, as a pre-#1706 file would be.
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	delete(doc, "announcedAlsoBlocks")
	if bf, ok := doc["battlefield"].(map[string]any); ok {
		for _, c := range bf["cards"].([]any) {
			delete(c.(map[string]any), "alsoBlocking")
		}
	}
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := blockedSet(restored, w); !sameIDs(got, []uuid.UUID{a}) {
		t.Fatalf("old-format block restored as %v", got)
	}
	restored.WithWriteLock(func() {
		if got := restored.announcedBlocks[w]; !sameIDs(got, []uuid.UUID{a}) {
			t.Errorf("old-format announcement restored as %v", got)
		}
		if restored.blockDeclarationPendingLocked() {
			t.Error("an old-format lock-in would announce again")
		}
	})
}
