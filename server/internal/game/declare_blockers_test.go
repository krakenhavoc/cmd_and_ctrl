package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// declare_blockers_test.go — #750, ADR 0045 addendum Decision 13: the
// block declaration is a SET, and the engine never holds a block it
// would refuse.
//
// The defect this closes: a block COUNT (menace's minimum of two,
// Hungering Hydra's maximum of one) is a property of a whole
// declaration, so a per-pair verb had nothing to judge. The lone
// block on a menace attacker was accepted, drawn, and logged, and
// only undone later — first at the damage step, then at the
// declaration's lock-in. Either way the defender had been told it was
// good. Now it is refused where it is made, with a reason they can
// read.

// menaceBoard is one menace attacker from seat 0 and `blockers`
// untapped creatures for seat 1, parked in the declare-blockers step.
func menaceBoard(t *testing.T, g *Game, blockers int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	attacker := pushCombatant(t, g, g.Seats[0], "Menacer", 3, 3, "menace")
	var bs []uuid.UUID
	for i := range blockers {
		bs = append(bs, pushCombatant(t, g, g.Seats[1], "Blocker "+string(rune('A'+i)), 1, 4))
	}
	declareAttacks(t, g, attacker)
	return attacker, bs
}

// blockingTargets reports what each named card is blocking, for the
// "the board is unchanged" assertions.
func blockingTargets(t *testing.T, g *Game, ids ...uuid.UUID) []uuid.UUID {
	t.Helper()
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		c := findCard(g, id)
		if c == nil {
			t.Fatalf("%s left the battlefield", id)
		}
		out = append(out, c.BlockingTarget)
	}
	return out
}

// A lone blocker on a menace attacker is refused AT DECLARATION, with
// a player-facing reason, and the board is untouched: this is the
// headline defect of #750.
func TestDeclareBlockersRefusesALoneMenaceBlockerAndChangesNothing(t *testing.T) {
	g := newActiveGame(t)
	attacker, bs := menaceBoard(t, g, 2)
	before := blockingTargets(t, g, bs[0], bs[1])
	seq := lastSeq(g)

	err := g.DeclareBlockers([]BlockDeclaration{{Blocker: bs[0], Attacker: attacker}})
	if !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("a lone menace blocker = %v, want an illegal-block refusal", err)
	}
	var refusal *BlockRefusedError
	if !errors.As(err, &refusal) {
		t.Fatalf("the refusal carries no reason: %v", err)
	}
	if refusal.Reason != BlockReasonTooFewBlockers {
		t.Errorf("reason = %q, want %q", refusal.Reason, BlockReasonTooFewBlockers)
	}
	if refusal.N != 2 {
		t.Errorf("N = %d, want the minimum, 2", refusal.N)
	}
	if refusal.Attacker != attacker {
		t.Errorf("the refusal names attacker %v, want %v", refusal.Attacker, attacker)
	}
	// The sentence is built server-side so the client never re-derives
	// a block rule (ADR 0045 §6).
	if got := refusal.Sentence(g.Seats[1].ID); got != "Menacer can't be blocked by fewer than two creatures." {
		t.Errorf("the defender reads %q", got)
	}

	// Nothing stored, nothing announced, nothing blocked.
	if got := blockingTargets(t, g, bs[0], bs[1]); got[0] != before[0] || got[1] != before[1] {
		t.Errorf("a refused declaration changed the board: %v, was %v", got, before)
	}
	if g.blockedAttackers[attacker] {
		t.Error("a refused declaration recorded the attacker as blocked")
	}
	for _, ev := range g.Events {
		if ev.Seq > seq && (ev.Kind == EventBlock || ev.Kind == EventBecomesBlocked) {
			t.Errorf("a refused declaration emitted %s", ev.Kind)
		}
	}
}

// The legal two-blocker menace set is accepted as one action.
func TestDeclareBlockersAcceptsALegalMenaceSet(t *testing.T) {
	g := newActiveGame(t)
	attacker, bs := menaceBoard(t, g, 2)

	if err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: bs[0], Attacker: attacker},
		{Blocker: bs[1], Attacker: attacker},
	}); err != nil {
		t.Fatalf("a legal two-creature menace block was refused: %v", err)
	}
	for _, got := range blockingTargets(t, g, bs[0], bs[1]) {
		if got != attacker {
			t.Fatalf("a legal menace block was not stored: blocking %v, want %v", got, attacker)
		}
	}
	g.WithWriteLock(func() { g.completeAllBlockDeclarationsLocked(); g.commitBlockDeclarationLocked() })
	if !g.blockedAttackers[attacker] {
		t.Error("the legal menace block did not make the attacker blocked")
	}
}

// All or nothing (Decision 13). A set whose SECOND entry is refused
// stores none of it — not even the first, legal entry. This is the
// difference from DeclareAttackers, which skips what it cannot take.
func TestDeclareBlockersIsAllOrNothing(t *testing.T) {
	// "Blocker C" can't block at all, so the set it is in must be
	// refused whole — including the two legal entries ahead of it.
	withRestrictionOn(t, CantBlock, "Blocker C")

	g := newActiveGame(t)
	pushRestrictor(g, g.Seats[0])
	attacker, bs := menaceBoard(t, g, 3)
	before := blockingTargets(t, g, bs[0], bs[1], bs[2])

	err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: bs[0], Attacker: attacker},
		{Blocker: bs[1], Attacker: attacker},
		{Blocker: bs[2], Attacker: attacker},
	})
	if !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("a set containing an illegal pair = %v, want a refusal", err)
	}
	got := blockingTargets(t, g, bs[0], bs[1], bs[2])
	for i := range got {
		if got[i] != before[i] {
			t.Fatalf("a refused set stored entry %d anyway", i)
		}
	}
}

// The single-blocker path is untouched for every ordinary attacker:
// one DeclareBlocker still stages one block, and is still idempotent.
func TestDeclareBlockerStillBlocksAnOrdinaryAttacker(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	blocker := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
	declareAttacks(t, g, attacker)

	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if got := blockingTargets(t, g, blocker)[0]; got != attacker {
		t.Fatalf("the block was not stored: blocking %v, want %v", got, attacker)
	}
	// Idempotent on the same pair: re-declaring is not a second
	// blocker and is not re-judged.
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("re-declaring the same pair: %v", err)
	}
	g.WithWriteLock(func() { g.completeAllBlockDeclarationsLocked(); g.commitBlockDeclarationLocked() })
	if n := len(blockDeclEvents(g, EventBlock, blocker)); n != 1 {
		t.Errorf("the block announced %d times, want 1", n)
	}

	// And the errors the per-pair verb has always returned still come
	// back through the set verb.
	if err := g.DeclareBlocker(uuid.New(), attacker); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("an unknown blocker = %v, want ErrCardNotFound", err)
	}
	if err := g.DeclareBlocker(blocker, uuid.New()); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("an unknown attacker = %v, want ErrCardNotFound", err)
	}
	if err := g.DeclareBlockers(nil); !errors.Is(err, ErrEmptyBlockerSet) {
		t.Errorf("an empty declaration = %v, want ErrEmptyBlockerSet", err)
	}
}

// Re-pointing a blocker out of a legal menace block would leave an
// illegal one behind, so the re-point is refused: Decision 13 checks
// every attacker whose blocker set the action CHANGES, including the
// one that LOSES a blocker.
func TestDeclareBlockersRefusesARepointThatBreaksAMenaceBlock(t *testing.T) {
	g := newActiveGame(t)
	menacer := pushCombatant(t, g, g.Seats[0], "Menacer", 3, 3, "menace")
	other := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	first := pushCombatant(t, g, g.Seats[1], "Blocker One", 1, 4)
	second := pushCombatant(t, g, g.Seats[1], "Blocker Two", 1, 4)
	declareAttacks(t, g, menacer, other)

	if err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: first, Attacker: menacer},
		{Blocker: second, Attacker: menacer},
	}); err != nil {
		t.Fatalf("a legal menace block was refused: %v", err)
	}
	err := g.DeclareBlocker(second, other)
	if !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("re-pointing half of a menace block = %v, want a refusal", err)
	}
	if got := blockingTargets(t, g, first, second); got[0] != menacer || got[1] != menacer {
		t.Errorf("the refused re-point disturbed the menace block: %v", got)
	}
}

// An attacker that GAINS menace after a legal block keeps that block:
// restrictions are checked as blockers are declared and never again
// (CR 509.1b, ADR 0045 §5). Only attackers the action touches are
// judged.
func TestDeclareBlockersDoesNotRejudgeAnUntouchedAttacker(t *testing.T) {
	g := newActiveGame(t)
	menacer := pushCombatant(t, g, g.Seats[0], "Late Menacer", 3, 3)
	other := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	first := pushCombatant(t, g, g.Seats[1], "Blocker One", 1, 4)
	second := pushCombatant(t, g, g.Seats[1], "Blocker Two", 1, 4)
	declareAttacks(t, g, menacer, other)

	if err := g.DeclareBlocker(first, menacer); err != nil {
		t.Fatalf("the block was legal when it was declared: %v", err)
	}
	// The attacker gains menace afterwards.
	g.WithWriteLock(func() {
		idx := findCardOnBattlefield(g, menacer)
		g.Battlefield.Cards[idx].Keywords = append(g.Battlefield.Cards[idx].Keywords, "menace")
		g.recomputeLayersLocked()
	})
	// A block declared on a DIFFERENT attacker does not re-judge it.
	if err := g.DeclareBlocker(second, other); err != nil {
		t.Fatalf("a block on an untouched attacker was refused: %v", err)
	}
	if got := blockingTargets(t, g, first)[0]; got != menacer {
		t.Error("an attacker that gained menace after a legal block lost it (CR 509.1b)")
	}
}

// The #328 auto-pass signal reads the same option generator the verb
// and the enumerator read, so it stops owing a decision the defender
// cannot act on: one untapped creature against a lone menace attacker
// is no legal block at all.
func TestSeatOwesNoBlockDecisionWithoutEnoughBlockersForMenace(t *testing.T) {
	g := newActiveGame(t)
	menaceBoard(t, g, 1)
	if g.SeatOwesBlockDecision(g.Seats[1].ID) {
		t.Error("one creature against a lone menace attacker is no legal block, so no decision is owed")
	}

	// A second untapped creature makes the two-creature block
	// available, and the decision is owed.
	g2 := newActiveGame(t)
	menaceBoard(t, g2, 2)
	if !g2.SeatOwesBlockDecision(g2.Seats[1].ID) {
		t.Error("two creatures CAN block a menace attacker, so the window must not be auto-passed")
	}

	// #1279: in the first game the defender had no legal block as the
	// step began, so their declaration — none — completed then
	// (CR 509.1). A creature that arrives afterwards does not reopen
	// it; before #1279 it did, because nothing recorded that the
	// declaration had happened.
	pushCombatant(t, g, g.Seats[1], "Blocker B", 1, 4)
	if g.SeatOwesBlockDecision(g.Seats[1].ID) {
		t.Error("a declaration completed at the step's start was reopened by a creature arriving after it")
	}
	if got := g.BlockDeclarationStatusOf(g.Seats[1].ID); got != BlockDeclarationDeclared {
		t.Errorf("status = %q, want declared", got)
	}
}

// BlockOptionsLocked is the one generator (Decision 14): it withholds
// the illegal single and offers the legal group.
func TestBlockOptionsOffersTheMenaceGroupAndNotTheSingle(t *testing.T) {
	g := newActiveGame(t)
	attacker, bs := menaceBoard(t, g, 2)

	var opts []BlockOption
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		opts = g.BlockOptionsLocked(g.Seats[1].ID, 12)
	})
	if len(opts) != 1 {
		t.Fatalf("got %d options, want exactly the one two-creature block", len(opts))
	}
	if len(opts[0].Blocks) != 2 {
		t.Fatalf("the option has %d entries, want 2", len(opts[0].Blocks))
	}
	for _, d := range opts[0].Blocks {
		if d.Attacker != attacker {
			t.Errorf("an entry names attacker %v, want %v", d.Attacker, attacker)
		}
	}
	if opts[0].Blocks[0].Blocker != bs[0] || opts[0].Blocks[1].Blocker != bs[1] {
		t.Error("the group is not in battlefield order")
	}
	// Every option the generator returns is one the verb accepts.
	if err := g.DeclareBlockers(opts[0].Blocks); err != nil {
		t.Fatalf("an offered option was refused by the verb: %v", err)
	}
}

// A maximum bounds the singles too: the generator offers the first
// blocker and, once it is stored, offers no second one.
func TestBlockOptionsRespectsAMaximum(t *testing.T) {
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key != blockRuleOracle {
			return nil
		}
		return []BlockRule{{
			Count: func(g *Game, attacker, source *Card) (int, int) { return 0, 1 },
		}}
	})

	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Hungering Hydra", 4, 4)
	first := pushCombatant(t, g, g.Seats[1], "Blocker One", 1, 4)
	pushCombatant(t, g, g.Seats[1], "Blocker Two", 1, 4)
	carryBlockRule(t, g, attacker)
	declareAttacks(t, g, attacker)

	count := func() int {
		var n int
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			n = len(g.BlockOptionsLocked(g.Seats[1].ID, 12))
		})
		return n
	}
	if got := count(); got != 2 {
		t.Fatalf("got %d options before any block, want one per creature", got)
	}
	if err := g.DeclareBlocker(first, attacker); err != nil {
		t.Fatalf("the first blocker is within the maximum: %v", err)
	}
	if got := count(); got != 0 {
		t.Errorf("got %d options once the maximum is reached, want 0", got)
	}
}

// TurnScopedBlockRules is an until-end-of-turn registry (Decision 11):
// its rules bind while the turn lasts and the cleanup sweep empties it.
func TestTurnScopedBlockRuleBindsThenSweepsAtEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Gingerbrute", 1, 1)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 2, 2)
	g.WithWriteLock(func() {
		g.RegisterTurnScopedBlockRuleLocked(BlockRule{
			Reason: BlockReasonCantBeBlockedExceptBy,
			Label:  "creatures with haste",
			Pair: func(g *Game, atk, blk, source *Card) bool {
				return atk != nil && atk.Effective().Name == "Gingerbrute"
			},
		})
	})
	if n := g.CaptureSnapshot().Continuations.TurnScopedBlockRules; n != 1 {
		t.Errorf("the census counts %d turn-scoped block rules, want 1", n)
	}
	declareAttacks(t, g, attacker)
	if err := g.DeclareBlocker(blocker, attacker); !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("a turn-scoped rule did not refuse the block: %v", err)
	}

	g.WithWriteLock(func() { g.ClearTurnScopedBlockRulesLocked() })
	if len(g.TurnScopedBlockRules) != 0 {
		t.Error("the cleanup sweep did not empty the registry")
	}
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Errorf("the swept rule still refuses the block: %v", err)
	}
}
