package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// combat_limits_test.go — #1507, CR 508.1c / 509.1b: the WHOLE-COMBAT
// count limits. Silent Arbiter's "no more than one creature can attack
// each combat" and "no more than one creature can block each combat",
// Crawlspace's "no more than two creatures can attack you each combat".
// ADR 0045 amendment of 2026-09-24, Decisions 43-45.
//
// The engine half only, with the catalog hooks stubbed; the cards, the
// legal-move enumerator and the bot are pinned in cards/effects and
// aiseat. What is pinned here is the contract both verbs share: a
// declaration that breaks a limit is refused WHOLE, nothing is staged,
// and a declaration that does not raise a limit's count is never
// refused by it.

const combatLimitOracle = "combat-limit-oracle-test"

// stubCombatLimits installs an attack-limit catalog and a block-rule
// catalog that answer only for combatLimitOracle, and restores both
// hooks afterwards.
func stubCombatLimits(t *testing.T, attack []AttackLimit, block []BlockRule) {
	t.Helper()
	prevA, prevB := CatalogAttackLimits, CatalogBlockRules
	t.Cleanup(func() { CatalogAttackLimits, CatalogBlockRules = prevA, prevB })
	CatalogAttackLimits = func(key string) []AttackLimit {
		if key == combatLimitOracle {
			return attack
		}
		return nil
	}
	CatalogBlockRules = func(key string) []BlockRule {
		if key == combatLimitOracle {
			return block
		}
		return nil
	}
}

// pushLimitSource puts a noncreature permanent carrying the stubbed
// limits onto the battlefield under `owner`.
func pushLimitSource(t *testing.T, g *Game, owner *Player, name string) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = "Artifact"
	c.OracleID = combatLimitOracle
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// arbiterBlockLimit is NoMoreThanNCanBlockEachCombat(1) without
// importing effects: every blocker counts toward one.
func arbiterBlockLimit(n int) BlockRule {
	return BlockRule{Limit: func(_ *Game, _, _ *Card) int { return n }}
}

func attackLimitErr(t *testing.T, err error) *AttackLimitError {
	t.Helper()
	if err == nil {
		t.Fatal("the declaration was accepted; want an attack-limit refusal")
	}
	if !errors.Is(err, ErrAttackLimit) {
		t.Fatalf("refusal does not wrap ErrAttackLimit: %v", err)
	}
	var le *AttackLimitError
	if !errors.As(err, &le) {
		t.Fatalf("refusal is %T, want *AttackLimitError", err)
	}
	return le
}

func attackingCount(g *Game) int {
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n
}

// --- attack side -------------------------------------------------

// TestAttackLimitRefusesTheSecondAttacker — Silent Arbiter's first
// line through the per-creature verb. The second creature is refused
// with the limit, the Arbiter and the number, and is left untapped
// and undeclared.
func TestAttackLimitRefusesTheSecondAttacker(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newActiveGame(t)
	arbiter := pushLimitSource(t, g, g.Seats[1], "Silent Arbiter")
	a := pushCombatant(t, g, g.Seats[0], "Bear A", 2, 2)
	b := pushCombatant(t, g, g.Seats[0], "Bear B", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("the first attacker is legal: %v", err)
	}
	le := attackLimitErr(t, g.DeclareAttacker(b, g.Seats[1].ID))
	if le.Max != 1 || le.Source != arbiter || le.SourceName != "Silent Arbiter" || le.Attacker != b {
		t.Errorf("refusal = %+v", le)
	}
	if want := "No more than one creature can attack each combat (Silent Arbiter)."; le.Sentence(uuid.Nil) != want {
		t.Errorf("sentence %q, want %q", le.Sentence(uuid.Nil), want)
	}
	if c := findCard(g, b); c.AttackingTarget != uuid.Nil || c.Tapped {
		t.Errorf("the refused creature was staged (target %v, tapped %v)", c.AttackingTarget, c.Tapped)
	}
	// Re-pointing the one attacker is not a second attacker.
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Errorf("re-declaring the standing attacker was refused: %v", err)
	}
}

// TestAttackLimitIsAllOrNothingOnTheBulkVerb — declare_attackers skips
// INELIGIBLE entries, but an over-full swing has none: which creature
// stays home is the player's choice (ADR 0080's reasoning), so the
// whole set is refused and nothing is tapped.
func TestAttackLimitIsAllOrNothingOnTheBulkVerb(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newActiveGame(t)
	pushLimitSource(t, g, g.Seats[1], "Silent Arbiter")
	a := pushCombatant(t, g, g.Seats[0], "Bear A", 2, 2)
	b := pushCombatant(t, g, g.Seats[0], "Bear B", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	_, err := g.DeclareAttackers([]AttackDeclaration{
		{Attacker: a, Target: g.Seats[1].ID}, {Attacker: b, Target: g.Seats[1].ID},
	})
	attackLimitErr(t, err)
	for _, id := range []uuid.UUID{a, b} {
		if c := findCard(g, id); c.AttackingTarget != uuid.Nil || c.Tapped {
			t.Errorf("%s was staged by a refused declaration", c.Name)
		}
	}
	declared, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: b, Target: g.Seats[1].ID}})
	if err != nil || len(declared) != 1 {
		t.Fatalf("a one-creature swing is legal: %v %v", declared, err)
	}
}

// TestAttackLimitAttackingYouCountsOnlyItsController — Crawlspace in a
// four-player game. Two creatures may attack its controller; a third is
// refused, and the other opponents may still be attacked by any number
// in the same combat. The sentence says "you" to the protected seat
// and names it to anybody else.
func TestAttackLimitAttackingYouCountsOnlyItsController(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitAttackingYou, Max: 2}}, nil)
	g := newFourPlayerActiveGame(t)
	me, crawl, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushLimitSource(t, g, crawl, "Crawlspace")
	var bears []uuid.UUID
	for i := 0; i < 5; i++ {
		bears = append(bears, pushCombatant(t, g, me, "Bear", 2, 2))
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	for _, b := range bears[:2] {
		if err := g.DeclareAttacker(b, crawl.ID); err != nil {
			t.Fatalf("two attackers at the Crawlspace player are legal: %v", err)
		}
	}
	le := attackLimitErr(t, g.DeclareAttacker(bears[2], crawl.ID))
	if le.Defender != crawl.ID || le.Max != 2 {
		t.Errorf("refusal = %+v", le)
	}
	if got := le.Sentence(crawl.ID); got != "No more than two creatures can attack you each combat (Crawlspace)." {
		t.Errorf("the protected seat reads %q", got)
	}
	if got := le.Sentence(me.ID); got != "No more than two creatures can attack "+crawl.Name+" each combat (Crawlspace)." {
		t.Errorf("the attacker reads %q", got)
	}
	// The limit is about ONE seat.
	for _, b := range bears[2:] {
		if err := g.DeclareAttacker(b, other.ID); err != nil {
			t.Fatalf("an attack on another opponent was refused: %v", err)
		}
	}
	// Re-pointing an attacker from the other opponent onto the
	// Crawlspace player RAISES that seat's count, so it is refused.
	attackLimitErr(t, g.DeclareAttacker(bears[4], crawl.ID))
	// …and moving one of the two away is never refused.
	if err := g.DeclareAttacker(bears[0], other.ID); err != nil {
		t.Fatalf("lowering the count was refused: %v", err)
	}
	if err := g.DeclareAttacker(bears[4], crawl.ID); err != nil {
		t.Fatalf("with a slot free the re-point is legal: %v", err)
	}
}

// TestAttackLimitAttackingYouIgnoresPlaneswalkers — "attack you" is the
// PLAYER, as ADR 0080 reads Propaganda's same word: attacks on a
// planeswalker the Crawlspace player controls are not counted.
func TestAttackLimitAttackingYouIgnoresPlaneswalkers(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitAttackingYou, Max: 1}}, nil)
	g := newActiveGame(t)
	me, crawl := g.Seats[0], g.Seats[1]
	pushLimitSource(t, g, crawl, "Judoon Enforcers")
	pw := pushPlaneswalkerForTest(g, crawl.ID, "Their Walker", 5)
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	c := pushCombatant(t, g, me, "Bear C", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(a, crawl.ID); err != nil {
		t.Fatalf("one attacker at the player: %v", err)
	}
	if err := g.DeclareAttacker(b, pw); err != nil {
		t.Fatalf("an attack on the planeswalker does not count: %v", err)
	}
	if err := g.DeclareAttacker(c, pw); err != nil {
		t.Fatalf("a second attack on the planeswalker does not count: %v", err)
	}
	if n := attackingCount(g); n != 3 {
		t.Errorf("%d attacking, want 3", n)
	}
}

// TestAttackLimitArrivingLateUnmakesNothing — a limit that arrives
// after the declaration (a flashed-in Arbiter) does not unmake the
// attackers standing: the verbs refuse only a declaration that RAISES
// the count.
func TestAttackLimitArrivingLateUnmakesNothing(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newFourPlayerActiveGame(t)
	a := pushCombatant(t, g, g.Seats[0], "Bear A", 2, 2)
	b := pushCombatant(t, g, g.Seats[0], "Bear B", 2, 2)
	c := pushCombatant(t, g, g.Seats[0], "Bear C", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("no limit yet: %v", err)
		}
	}
	pushLimitSource(t, g, g.Seats[1], "Silent Arbiter")

	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Errorf("re-pointing a standing attacker raises no count: %v", err)
	}
	attackLimitErr(t, g.DeclareAttacker(c, g.Seats[1].ID))
	if n := attackingCount(g); n != 2 {
		t.Errorf("%d attacking, want the 2 declared before the limit", n)
	}
}

// --- block side --------------------------------------------------

func blockLimitRefusal(t *testing.T, err error) *BlockRefusedError {
	t.Helper()
	var refused *BlockRefusedError
	if err == nil || !errors.As(err, &refused) || refused.Reason != BlockReasonDeclarationLimit {
		t.Fatalf("want a declaration_limit refusal, got %v", err)
	}
	return refused
}

// TestBlockLimitCountsEveryBlockInTheCombat — Silent Arbiter's second
// line. One creature may block, on any attacker; a second is refused
// with declaration_limit whichever attacker it names, as a single or
// inside a set, and the option generator stops offering blocks once
// the one is used.
func TestBlockLimitCountsEveryBlockInTheCombat(t *testing.T) {
	stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(1)})
	g := newActiveGame(t)
	arbiter := pushLimitSource(t, g, g.Seats[1], "Silent Arbiter")
	atkA := pushCombatant(t, g, g.Seats[0], "Attacker A", 2, 2)
	atkB := pushCombatant(t, g, g.Seats[0], "Attacker B", 2, 2)
	blkA := pushCombatant(t, g, g.Seats[1], "Blocker A", 2, 2)
	blkB := pushCombatant(t, g, g.Seats[1], "Blocker B", 2, 2)
	declareAttacks(t, g, atkA, atkB)

	// Two blocks at once: refused whole.
	refused := blockLimitRefusal(t, g.DeclareBlockers([]BlockDeclaration{
		{Blocker: blkA, Attacker: atkA}, {Blocker: blkB, Attacker: atkB},
	}))
	if refused.N != 1 || refused.Source != arbiter || refused.SourceName != "Silent Arbiter" {
		t.Errorf("refusal = %+v", refused)
	}
	if want := "No more than one creature can block each combat (Silent Arbiter)."; refused.Sentence(uuid.Nil) != want {
		t.Errorf("sentence %q, want %q", refused.Sentence(uuid.Nil), want)
	}
	if findCard(g, blkA).BlockingTarget != uuid.Nil {
		t.Error("a refused declaration staged a block")
	}

	offered := len(g.BlockOptionsLocked(g.Seats[1].ID, 4))
	if offered != 4 {
		t.Errorf("before any block, %d singles are offered, want 4", offered)
	}
	if err := g.DeclareBlocker(blkA, atkA); err != nil {
		t.Fatalf("one block is legal: %v", err)
	}
	blockLimitRefusal(t, g.DeclareBlocker(blkB, atkB))
	blockLimitRefusal(t, g.DeclareBlocker(blkB, atkA))
	if opts := g.BlockOptionsLocked(g.Seats[1].ID, 4); len(opts) != 0 {
		t.Errorf("the generator offers %d blocks past the limit", len(opts))
	}
	// Re-pointing the one blocker raises no count.
	if err := g.DeclareBlocker(blkA, atkB); err != nil {
		t.Errorf("re-pointing the one blocker was refused: %v", err)
	}
}

// TestBlockLimitArrivingLateUnmakesNothing — the block side of the
// same rule: two blocks stand before the limit appears, re-pointing one
// of them raises no count and is accepted, and a third blocker is
// refused.
func TestBlockLimitArrivingLateUnmakesNothing(t *testing.T) {
	stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(1)})
	g := newActiveGame(t)
	atkA := pushCombatant(t, g, g.Seats[0], "Attacker A", 2, 2)
	atkB := pushCombatant(t, g, g.Seats[0], "Attacker B", 2, 2)
	blkA := pushCombatant(t, g, g.Seats[1], "Blocker A", 2, 2)
	blkB := pushCombatant(t, g, g.Seats[1], "Blocker B", 2, 2)
	blkC := pushCombatant(t, g, g.Seats[1], "Blocker C", 2, 2)
	declareAttacks(t, g, atkA, atkB)
	if err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: blkA, Attacker: atkA}, {Blocker: blkB, Attacker: atkA},
	}); err != nil {
		t.Fatalf("no limit yet: %v", err)
	}
	pushLimitSource(t, g, g.Seats[1], "Silent Arbiter")

	if err := g.DeclareBlocker(blkB, atkB); err != nil {
		t.Errorf("re-pointing a standing blocker raises no count: %v", err)
	}
	blockLimitRefusal(t, g.DeclareBlocker(blkC, atkB))
}

// TestBlockLimitSpansDefenders — multiplayer: every stored block in
// the combat counts, whichever defender made it. The first defender to
// block uses the one up, and the second defender is offered nothing.
func TestBlockLimitSpansDefenders(t *testing.T) {
	stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(1)})
	g := newFourPlayerActiveGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	pushLimitSource(t, g, me, "Silent Arbiter")
	atk1 := pushCombatant(t, g, me, "Attacker 1", 2, 2)
	atk2 := pushCombatant(t, g, me, "Attacker 2", 2, 2)
	blk1 := pushCombatant(t, g, p1, "Blocker 1", 2, 2)
	blk2 := pushCombatant(t, g, p2, "Blocker 2", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(atk1, p1.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(atk2, p2.ID); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	if err := g.DeclareBlocker(blk1, atk1); err != nil {
		t.Fatalf("the first defender's block: %v", err)
	}
	blockLimitRefusal(t, g.DeclareBlocker(blk2, atk2))
	if opts := g.BlockOptionsLocked(p2.ID, 4); len(opts) != 0 {
		t.Errorf("the second defender is offered %d blocks past the limit", len(opts))
	}
	if g.SeatOwesBlockDecision(p2.ID) {
		t.Error("the #328 signal holds the second defender for a block it cannot make")
	}
}

// TestBlockLimitAndMenace — a menace attacker needs two blockers and
// the combat allows one, so it is unblockable: the singles are
// too_few_blockers, the pair is declaration_limit, and nothing is
// offered.
func TestBlockLimitAndMenace(t *testing.T) {
	stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(1)})
	g := newActiveGame(t)
	pushLimitSource(t, g, g.Seats[1], "Silent Arbiter")
	atk := pushCombatant(t, g, g.Seats[0], "Menace", 3, 3, "menace")
	blkA := pushCombatant(t, g, g.Seats[1], "Blocker A", 2, 2)
	blkB := pushCombatant(t, g, g.Seats[1], "Blocker B", 2, 2)
	declareAttacks(t, g, atk)

	blockLimitRefusal(t, g.DeclareBlockers([]BlockDeclaration{
		{Blocker: blkA, Attacker: atk}, {Blocker: blkB, Attacker: atk},
	}))
	if opts := g.BlockOptionsLocked(g.Seats[1].ID, 4); len(opts) != 0 {
		t.Errorf("%d blocks offered on a menace attacker under a one-blocker combat", len(opts))
	}
}

// TestBlockLimitTheTightestWins — two limits are judged separately, so
// a bound of two beside a bound of one allows one.
func TestBlockLimitTheTightestWins(t *testing.T) {
	stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(2), arbiterBlockLimit(1)})
	g := newActiveGame(t)
	pushLimitSource(t, g, g.Seats[1], "Caverns and Arbiter")
	atk := pushCombatant(t, g, g.Seats[0], "Attacker", 5, 5)
	blkA := pushCombatant(t, g, g.Seats[1], "Blocker A", 2, 2)
	blkB := pushCombatant(t, g, g.Seats[1], "Blocker B", 2, 2)
	declareAttacks(t, g, atk)

	refused := blockLimitRefusal(t, g.DeclareBlockers([]BlockDeclaration{
		{Blocker: blkA, Attacker: atk}, {Blocker: blkB, Attacker: atk},
	}))
	if refused.N != 1 {
		t.Errorf("the refusal names bound %d, want the tighter 1", refused.N)
	}
}
