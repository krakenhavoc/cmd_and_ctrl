package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// clAgreeOnBlocks asserts every single block the verb accepts is
// offered and every offered block is accepted, and returns how many
// block moves were offered.
func clAgreeOnBlocks(t *testing.T, g *game.Game, seat uuid.UUID) int {
	t.Helper()
	offered := effectstest.BrOffered(t, g, seat)
	for _, set := range offered {
		if err := g.Clone().DeclareBlockers(set); err != nil {
			t.Fatalf("the enumerator offers %v, which the verb refuses: %v", set, err)
		}
	}
	for _, b := range g.Battlefield.Cards {
		if b.Controller != seat || !b.IsCreature() || b.Tapped || b.BlockingTarget != uuid.Nil {
			continue
		}
		for _, a := range g.Battlefield.Cards {
			if a.AttackingTarget == uuid.Nil {
				continue
			}
			if g.Clone().DeclareBlocker(b.InstanceID, a.InstanceID) == nil && !effectstest.BrOffers(offered, b.InstanceID, a.InstanceID) {
				t.Fatalf("the verb accepts %s blocking %s but the enumerator withholds it", b.Name, a.Name)
			}
		}
	}
	return len(offered)
}

// --- Silent Arbiter ---------------------------------------------

// Attack side: every creature is offered until one attacks, then none
// is, and the verb refuses the second with the Arbiter named.
func TestSilentArbiterAllowsOneAttackerEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbiter := effectstest.B12Push(g, opp.ID, "Silent Arbiter", "Artifact Creature — Construct", effectstest.SilentArbiterOracle, 1, 5)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)

	if n := len(effectstest.ClAgreeOnAttacks(t, g, me.ID)); n != 6 {
		t.Fatalf("before any attack, %d attacks are offered; want two creatures at three opponents", n)
	}
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("the first attacker: %v", err)
	}
	if n := len(effectstest.ClAgreeOnAttacks(t, g, me.ID)); n != 0 {
		t.Errorf("after one attacker, %d more attacks are offered", n)
	}
	le := effectstest.ClAttackRefusal(t, g.DeclareAttacker(b, g.Seats[2].ID))
	if le.Source != arbiter || le.Max != 1 {
		t.Errorf("refusal = %+v", le)
	}
}

// Block side: one blocker in the whole combat, across attackers — and
// the #328 signal lets the defender go once it is used.
func TestSilentArbiterAllowsOneBlockerEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Silent Arbiter", "Artifact Creature — Construct", effectstest.SilentArbiterOracle, 1, 5)
	atk := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	blkA := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	blkB := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, atk)

	if n := clAgreeOnBlocks(t, g, opp.ID); n != 2 {
		t.Fatalf("before any block, %d blocks are offered, want 2", n)
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blkA, Attacker: atk}, {Blocker: blkB, Attacker: atk},
	}), game.BlockReasonDeclarationLimit)
	if refused.SourceName != "Silent Arbiter" || refused.N != 1 {
		t.Errorf("refusal = %+v", refused)
	}
	if err := g.DeclareBlocker(blkA, atk); err != nil {
		t.Fatalf("one block is legal: %v", err)
	}
	if n := clAgreeOnBlocks(t, g, opp.ID); n != 0 {
		t.Errorf("after one block, %d more are offered", n)
	}
	if g.SeatOwesBlockDecision(opp.ID) || effectstest.BrViewOwes(g) {
		t.Error("block_decision_seats holds the defender for a block it cannot make")
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(blkB, atk), game.BlockReasonDeclarationLimit)
}

// CR 613.1f: an Arbiter that has lost its abilities limits nothing.
func TestSilentArbiterWithNoAbilitiesLimitsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbiter := effectstest.B12Push(g, opp.ID, "Silent Arbiter", "Artifact Creature — Construct", effectstest.SilentArbiterOracle, 1, 5)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(arbiter),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("an Arbiter with no abilities refused an attacker: %v", err)
		}
	}
}

// --- Dueling Grounds ---------------------------------------------

func TestDuelingGroundsAllowsOneAttackerAndOneBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Dueling Grounds", "Enchantment", effectstest.DuelingGroundsOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	blkA := effectstest.B12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
	blkB := effectstest.B12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)

	// Its own controller is bound too: the line has no "you".
	_, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a, Target: opp.ID}, {Attacker: b, Target: opp.ID}})
	effectstest.ClAttackRefusal(t, err)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a, Target: opp.ID}}); err != nil {
		t.Fatalf("a one-creature swing: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blkA, a); err != nil {
		t.Fatalf("the one blocker: %v", err)
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(blkB, a), game.BlockReasonDeclarationLimit)
}

// --- Caverns of Despair -------------------------------------------

func TestCavernsOfDespairAllowsTwoAttackersAndTwoBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Caverns of Despair", "World Enchantment", effectstest.CavernsOfDespairOracle, 0, 0)
	var atk []uuid.UUID
	for i := 0; i < 3; i++ {
		atk = append(atk, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	var blk []uuid.UUID
	for i := 0; i < 3; i++ {
		blk = append(blk, effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range atk[:2] {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("two attackers are legal: %v", err)
		}
	}
	if n := len(effectstest.ClAgreeOnAttacks(t, g, me.ID)); n != 0 {
		t.Errorf("%d attacks offered past two", n)
	}
	if le := effectstest.ClAttackRefusal(t, g.DeclareAttacker(atk[2], opp.ID)); le.Max != 2 {
		t.Errorf("refusal names bound %d, want 2", le.Max)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blk[0], Attacker: atk[0]}, {Blocker: blk[1], Attacker: atk[1]},
	}); err != nil {
		t.Fatalf("two blockers are legal: %v", err)
	}
	if n := clAgreeOnBlocks(t, g, opp.ID); n != 0 {
		t.Errorf("%d blocks offered past two", n)
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(blk[2], atk[0]), game.BlockReasonDeclarationLimit)
}

// --- Crawlspace ---------------------------------------------------

// Multiplayer: two creatures may attack the Crawlspace player; the
// enumerator stops offering that seat at two and keeps offering the
// others, and the verb agrees.
func TestCrawlspaceLimitsAttackersPerDefendingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, crawl, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, crawl.ID, "Crawlspace", "Artifact", effectstest.CrawlspaceOracle, 0, 0)
	var bears []uuid.UUID
	for i := 0; i < 4; i++ {
		bears = append(bears, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)

	if at := effectstest.ClTargets(effectstest.ClAgreeOnAttacks(t, g, me.ID)); at[crawl.ID] != 4 || at[other.ID] != 4 {
		t.Fatalf("before any attack every creature is offered at every seat: %v", at)
	}
	for _, id := range bears[:2] {
		if err := g.DeclareAttacker(id, crawl.ID); err != nil {
			t.Fatalf("two attackers at the Crawlspace player: %v", err)
		}
	}
	at := effectstest.ClTargets(effectstest.ClAgreeOnAttacks(t, g, me.ID))
	if at[crawl.ID] != 0 {
		t.Errorf("%d attacks at the Crawlspace player offered past two", at[crawl.ID])
	}
	if at[other.ID] != 2 || at[g.Seats[3].ID] != 2 {
		t.Errorf("the other opponents stopped being offered: %v", at)
	}
	le := effectstest.ClAttackRefusal(t, g.DeclareAttacker(bears[2], crawl.ID))
	if le.Defender != crawl.ID || le.SourceName != "Crawlspace" {
		t.Errorf("refusal = %+v", le)
	}
	for _, id := range bears[2:] {
		if err := g.DeclareAttacker(id, other.ID); err != nil {
			t.Fatalf("another opponent may still be attacked: %v", err)
		}
	}
}

// "You" is the Crawlspace's controller: it does not limit attacks on
// its controller's opponents when its controller is the one attacking.
func TestCrawlspaceDoesNotLimitItsControllersAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Crawlspace", "Artifact", effectstest.CrawlspaceOracle, 0, 0)
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range bears {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("my own Crawlspace limited my attack: %v", err)
		}
	}
}

// --- Judoon Enforcers ---------------------------------------------

func TestJudoonEnforcersAllowsOneAttackerAtItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, judoon, other := g.Seats[0], g.Seats[1], g.Seats[2]
	enforcers := effectstest.B12Push(g, judoon.ID, "Judoon Enforcers", "Creature — Alien Rhino Soldier", effectstest.JudoonEnforcersOracle, 8, 8)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	if !game.HasKeyword(effectstest.MustBattlefieldCard(t, g, enforcers), "trample") {
		t.Error("Judoon Enforcers has trample")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, judoon.ID); err != nil {
		t.Fatalf("one attacker: %v", err)
	}
	if at := effectstest.ClTargets(effectstest.ClAgreeOnAttacks(t, g, me.ID)); at[judoon.ID] != 0 || at[other.ID] != 1 {
		t.Errorf("offers after one attacker: %v", at)
	}
	if le := effectstest.ClAttackRefusal(t, g.DeclareAttacker(b, judoon.ID)); le.Max != 1 {
		t.Errorf("refusal = %+v", le)
	}
	if err := g.DeclareAttacker(b, other.ID); err != nil {
		t.Fatalf("another opponent: %v", err)
	}
}

func TestJudoonEnforcersCanBeSuspended(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.JudoonEnforcersOracle)
	if !ok {
		t.Fatal("Judoon Enforcers is not registered")
	}
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Kind != game.SpecialActionSuspend ||
		spec.SpecialActions[0].Counters != 6 || spec.SpecialActions[0].Cost != "{1}{R}{W}" {
		t.Errorf("suspend = %+v, want Suspend 6—{1}{R}{W}", spec.SpecialActions)
	}
}
