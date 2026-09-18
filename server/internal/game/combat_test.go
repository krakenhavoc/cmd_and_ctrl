package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// combat_test.go covers the S18 sub-PR 3 combat damage rewrite:
// two-substep flow (CR 510.2 first-strike + CR 510.3 regular),
// deathtouch flag, lifelink life gain, trample overflow, menace
// revert, and the multi-blocker damage-assignment prompt.

// pushKeywordCreature drops a creature on the battlefield with the
// given keywords baked into its effective characteristic. Bypasses
// the real layer engine so tests don't need catalog entries.
func pushKeywordCreature(t *testing.T, g *Game, owner *Player, power, toughness int, keywords ...string) uuid.UUID {
	t.Helper()
	c := NewCard("Kw Creature", owner.ID)
	c.TypeLine = "Creature — Test"
	c.Power = power
	c.Toughness = toughness
	g.Battlefield.PushTop(c)
	// Set effective directly; these tests don't exercise the layer
	// engine since the catalog-integration path lands with sub-PR 4.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == c.InstanceID {
			g.Battlefield.Cards[i].effective = printedEffectiveWith(c, keywords...)
		}
	}
	return c.InstanceID
}

func findCard(g *Game, id uuid.UUID) *Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// advanceIntoCombatDamage steps into and past declare-attackers /
// declare-blockers, letting tests register attack + block at the
// right moments. Usage: advanceIntoStep(t, g, StepDeclareAttackers),
// declare, advanceIntoStep(t, g, StepDeclareBlockers), block,
// advanceIntoStep(t, g, StepCombatDamage).
func advanceIntoStep(t *testing.T, g *Game, step Step) {
	t.Helper()
	for g.Turn.Step != step {
		before := g.Turn
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to %v: %v", step, err)
		}
		if g.Turn == before {
			t.Fatalf("advanceIntoStep stuck at %v", before)
		}
	}
}

func TestCombatLifelinkGainsLife(t *testing.T) {
	g := newActiveGame(t)
	attackerID := pushKeywordCreature(t, g, g.Seats[0], 3, 3, "lifelink")
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attackerID, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	startLife := g.Seats[0].Life
	advanceIntoStep(t, g, StepCombatDamage)
	if g.Seats[0].Life-startLife != 3 {
		t.Errorf("lifelink gained %d life, want 3", g.Seats[0].Life-startLife)
	}
	if g.Seats[1].Life != 40-3 {
		t.Errorf("defender took %d life, want 3 of 40", 40-g.Seats[1].Life)
	}
}

func TestCombatDeathtouchDestroys5Over5Blocker(t *testing.T) {
	g := newActiveGame(t)
	rat := pushKeywordCreature(t, g, g.Seats[0], 1, 1, "deathtouch")
	ogre := pushKeywordCreature(t, g, g.Seats[1], 5, 5)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(rat, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(ogre, rat); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	// The 5/5 ogre should be gone (flagged by deathtouch → SBA).
	if findCard(g, ogre) != nil {
		t.Errorf("ogre should be destroyed by deathtouch, still on battlefield")
	}
	// The 1/1 rat dies to the ogre's 5 damage.
	if findCard(g, rat) != nil {
		t.Errorf("rat should die to ogre's damage, still on battlefield")
	}
}

func TestCombatTrampleOverflowsToPlayer(t *testing.T) {
	g := newActiveGame(t)
	wurm := pushKeywordCreature(t, g, g.Seats[0], 6, 6, "trample")
	bear := pushKeywordCreature(t, g, g.Seats[1], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(wurm, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, wurm); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	// 3 lethal to bear, 3 trample to defender.
	if findCard(g, bear) != nil {
		t.Errorf("bear should be destroyed")
	}
	if g.Seats[1].Life != 40-3 {
		t.Errorf("defender life = %d, want 37 (3 trample)", g.Seats[1].Life)
	}
	// Wurm takes bear's 3 damage; 6 toughness > 3, survives.
	w := findCard(g, wurm)
	if w == nil || w.DamageMarked != 3 {
		t.Errorf("wurm damage = %d, want 3", w.DamageMarked)
	}
}

func TestCombatTrampleNoOverflowIfBlockerSurvives(t *testing.T) {
	g := newActiveGame(t)
	// 3/3 trample vs 5/5 blocker — no overflow because blocker isn't
	// at-least-lethal.
	dreadmaw := pushKeywordCreature(t, g, g.Seats[0], 3, 3, "trample")
	ogre := pushKeywordCreature(t, g, g.Seats[1], 5, 5)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(dreadmaw, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(ogre, dreadmaw)
	advanceIntoStep(t, g, StepCombatDamage)
	if g.Seats[1].Life != 40 {
		t.Errorf("defender took %d damage, want 0 (blocker didn't die)", 40-g.Seats[1].Life)
	}
}

func TestCombatFirstStrikeKillsBeforeRegular(t *testing.T) {
	g := newActiveGame(t)
	knight := pushKeywordCreature(t, g, g.Seats[0], 2, 2, "first strike")
	bear := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(knight, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(bear, knight)
	advanceIntoStep(t, g, StepCombatDamage)
	// Bear dies to first-strike damage before it can hit back.
	if findCard(g, bear) != nil {
		t.Errorf("bear should die to first-strike damage")
	}
	// Knight survives with 0 damage marked (bear was dead before
	// regular substep).
	k := findCard(g, knight)
	if k == nil {
		t.Fatalf("knight should survive")
	}
	if k.DamageMarked != 0 {
		t.Errorf("knight damage = %d, want 0 (bear dead before regular substep)", k.DamageMarked)
	}
}

func TestCombatDoubleStrikeHitsTwice(t *testing.T) {
	g := newActiveGame(t)
	// 1/1 double strike vs 2/2 vanilla → DS deals 1 first-strike;
	// bear survives with 1 damage; DS deals 1 regular; bear dies.
	// Bear's regular damage still lands on the DS (simultaneous
	// regular substep), so DS takes 2 → dies.
	ds := pushKeywordCreature(t, g, g.Seats[0], 1, 1, "double strike")
	bear := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(ds, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(bear, ds)
	advanceIntoStep(t, g, StepCombatDamage)
	if findCard(g, bear) != nil {
		t.Errorf("bear should be dead (1 first-strike + 1 regular = 2)")
	}
	if findCard(g, ds) != nil {
		t.Errorf("double-strike 1/1 should die to bear's regular 2 damage")
	}
}

func TestCombatMenaceRevertsSingleBlocker(t *testing.T) {
	g := newActiveGame(t)
	rogue := pushKeywordCreature(t, g, g.Seats[0], 3, 3, "menace")
	bear := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(rogue, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(bear, rogue)
	advanceIntoStep(t, g, StepCombatDamage)
	// Menace with 1 blocker → blocker reverted; attacker unblocked →
	// defender takes full damage.
	if g.Seats[1].Life != 40-3 {
		t.Errorf("defender life = %d, want 37 (menace unblocked)", g.Seats[1].Life)
	}
	b := findCard(g, bear)
	if b != nil && b.DamageMarked != 0 {
		t.Errorf("bear damage = %d, want 0 (never actually blocked)", b.DamageMarked)
	}
}

func TestCombatMenaceAcceptsTwoBlockers(t *testing.T) {
	g := newActiveGame(t)
	rogue := pushKeywordCreature(t, g, g.Seats[0], 3, 3, "menace")
	bear1 := pushKeywordCreature(t, g, g.Seats[1], 1, 1)
	bear2 := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(rogue, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(bear1, rogue)
	g.DeclareBlocker(bear2, rogue)
	// Advancing to combat damage queues the multi-blocker prompt
	// since attacker power=3 > 1 blocker; the server pauses.
	advanceIntoStep(t, g, StepCombatDamage)
	// A damage-assignment prompt should now exist for the attacker's
	// controller.
	var prompt *PendingChoice
	for _, pc := range g.PendingChoices {
		if pc.Kind == PendingChoiceDamageAssignment {
			prompt = pc
			break
		}
	}
	if prompt == nil {
		t.Fatalf("expected damage_assignment prompt after menace-valid block")
	}
	if prompt.Chooser != g.Seats[0].ID {
		t.Errorf("prompt chooser = %v, want attacker controller seat 0", prompt.Chooser)
	}
}

func TestResolveDamageAssignmentPrefixLethal(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 5, 5)
	bear1 := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	bear2 := pushKeywordCreature(t, g, g.Seats[1], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(attacker, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(bear1, attacker)
	g.DeclareBlocker(bear2, attacker)
	advanceIntoStep(t, g, StepCombatDamage)
	var promptID uuid.UUID
	for _, pc := range g.PendingChoices {
		if pc.Kind == PendingChoiceDamageAssignment {
			promptID = pc.ID
			break
		}
	}
	if promptID == uuid.Nil {
		t.Fatalf("expected damage_assignment prompt")
	}
	// Reject non-lethal-prefix: [bear1=1, bear2=4] — bear1 not at-least-lethal
	// but bear2 still got damage.
	err := g.ResolveDamageAssignment(promptID, g.Seats[0].ID,
		[]DamageAssignmentEntry{
			{BlockerID: bear1, Amount: 1},
			{BlockerID: bear2, Amount: 4},
		}, 0)
	if err != ErrInvalidParam {
		t.Errorf("non-lethal prefix: got %v, want ErrInvalidParam", err)
	}
	// Accept [bear1=2, bear2=3] — both lethal.
	err = g.ResolveDamageAssignment(promptID, g.Seats[0].ID,
		[]DamageAssignmentEntry{
			{BlockerID: bear1, Amount: 2},
			{BlockerID: bear2, Amount: 3},
		}, 0)
	if err != nil {
		t.Errorf("lethal split: got %v, want nil", err)
	}
	if findCard(g, bear1) != nil || findCard(g, bear2) != nil {
		t.Errorf("both blockers should be destroyed after lethal assignment")
	}
}

func TestResolveDamageAssignmentTrampleOverflow(t *testing.T) {
	g := newActiveGame(t)
	dreadmaw := pushKeywordCreature(t, g, g.Seats[0], 6, 6, "trample")
	bear1 := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	bear2 := pushKeywordCreature(t, g, g.Seats[1], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(dreadmaw, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(bear1, dreadmaw)
	g.DeclareBlocker(bear2, dreadmaw)
	advanceIntoStep(t, g, StepCombatDamage)
	var promptID uuid.UUID
	for _, pc := range g.PendingChoices {
		if pc.Kind == PendingChoiceDamageAssignment {
			promptID = pc.ID
			break
		}
	}
	if promptID == uuid.Nil {
		t.Fatalf("expected damage_assignment prompt")
	}
	// 2 to bear1 (lethal), 3 to bear2 (lethal), 1 trample to player.
	err := g.ResolveDamageAssignment(promptID, g.Seats[0].ID,
		[]DamageAssignmentEntry{
			{BlockerID: bear1, Amount: 2},
			{BlockerID: bear2, Amount: 3},
		}, 1)
	if err != nil {
		t.Errorf("trample split: got %v, want nil", err)
	}
	if g.Seats[1].Life != 40-1 {
		t.Errorf("defender life = %d, want 39", g.Seats[1].Life)
	}
}

func TestResolveDamageAssignmentRejectsTrampleWithoutKeyword(t *testing.T) {
	g := newActiveGame(t)
	// Vanilla 6/6, no trample.
	attacker := pushKeywordCreature(t, g, g.Seats[0], 6, 6)
	bear1 := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	bear2 := pushKeywordCreature(t, g, g.Seats[1], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(attacker, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(bear1, attacker)
	g.DeclareBlocker(bear2, attacker)
	advanceIntoStep(t, g, StepCombatDamage)
	var promptID uuid.UUID
	for _, pc := range g.PendingChoices {
		if pc.Kind == PendingChoiceDamageAssignment {
			promptID = pc.ID
			break
		}
	}
	if promptID == uuid.Nil {
		t.Fatalf("expected damage_assignment prompt")
	}
	// Reject trample attempt on non-trampler.
	err := g.ResolveDamageAssignment(promptID, g.Seats[0].ID,
		[]DamageAssignmentEntry{
			{BlockerID: bear1, Amount: 2},
			{BlockerID: bear2, Amount: 3},
		}, 1)
	if err != ErrInvalidParam {
		t.Errorf("trample without keyword: got %v, want ErrInvalidParam", err)
	}
}

func TestResolveDamageAssignmentDeathtouchOneIsLethal(t *testing.T) {
	g := newActiveGame(t)
	// 5/5 deathtouch+trample vs two 3/3 bears: assign 1+1=2 (deathtouch
	// lethal), spill 3 to player.
	attacker := pushKeywordCreature(t, g, g.Seats[0], 5, 5, "deathtouch", "trample")
	bear1 := pushKeywordCreature(t, g, g.Seats[1], 3, 3)
	bear2 := pushKeywordCreature(t, g, g.Seats[1], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(attacker, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(bear1, attacker)
	g.DeclareBlocker(bear2, attacker)
	advanceIntoStep(t, g, StepCombatDamage)
	var promptID uuid.UUID
	for _, pc := range g.PendingChoices {
		if pc.Kind == PendingChoiceDamageAssignment {
			promptID = pc.ID
			break
		}
	}
	if promptID == uuid.Nil {
		t.Fatalf("expected damage_assignment prompt")
	}
	err := g.ResolveDamageAssignment(promptID, g.Seats[0].ID,
		[]DamageAssignmentEntry{
			{BlockerID: bear1, Amount: 1},
			{BlockerID: bear2, Amount: 1},
		}, 3)
	if err != nil {
		t.Errorf("deathtouch+trample split: got %v, want nil", err)
	}
	// Both bears destroyed by deathtouch flag; player took 3.
	if findCard(g, bear1) != nil || findCard(g, bear2) != nil {
		t.Errorf("both bears should be destroyed by deathtouch")
	}
	if g.Seats[1].Life != 40-3 {
		t.Errorf("defender life = %d, want 37", g.Seats[1].Life)
	}
}

// TestDeclareAttackerTapsWithoutVigilance verifies the CR 508.1f cost:
// a creature without vigilance is tapped when declared as an attacker.
// Regression for the post-S18 bug where DeclareAttacker set
// AttackingTarget but never tapped the card — the vigilance-skip case
// looked correct only because the base case also did nothing.
func TestDeclareAttackerTapsWithoutVigilance(t *testing.T) {
	g := newActiveGame(t)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	b := findCard(g, bear)
	if b == nil {
		t.Fatalf("bear missing after declare")
	}
	if !b.Tapped {
		t.Errorf("attacker without vigilance should be tapped; Tapped=%v", b.Tapped)
	}
}

// TestDeclareAttackerVigilanceSkipsTap verifies CR 702.20: a creature
// with vigilance does NOT tap when declared as an attacker.
func TestDeclareAttackerVigilanceSkipsTap(t *testing.T) {
	g := newActiveGame(t)
	angel := pushKeywordCreature(t, g, g.Seats[0], 4, 4, "vigilance")
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(angel, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	a := findCard(g, angel)
	if a == nil {
		t.Fatalf("angel missing after declare")
	}
	if a.Tapped {
		t.Errorf("vigilance attacker should stay untapped; Tapped=%v", a.Tapped)
	}
}

// TestDeclareBlockerFlyingRejectsGroundBlocker verifies the CR 509.1b
// evasion gate: a vanilla ground creature cannot block a flyer.
// Regression for the post-S18 bug where DeclareBlocker never called
// CanBlock, so any creature could block anything.
func TestDeclareBlockerFlyingRejectsGroundBlocker(t *testing.T) {
	g := newActiveGame(t)
	flyer := pushKeywordCreature(t, g, g.Seats[0], 2, 2, "flying")
	ground := pushKeywordCreature(t, g, g.Seats[1], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(flyer, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(ground, flyer); !errors.Is(err, ErrIllegalBlock) {
		t.Errorf("ground vs flying: got %v, want ErrIllegalBlock", err)
	}
}

// TestDeclareBlockerReachCanBlockFlying verifies reach passes the
// CanBlock gate for flyers.
func TestDeclareBlockerReachCanBlockFlying(t *testing.T) {
	g := newActiveGame(t)
	flyer := pushKeywordCreature(t, g, g.Seats[0], 2, 2, "flying")
	spider := pushKeywordCreature(t, g, g.Seats[1], 2, 4, "reach")
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(flyer, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(spider, flyer); err != nil {
		t.Errorf("reach should block flyer: %v", err)
	}
}

// TestDeclareBlockerFlyingCanBlockFlying verifies flyer-on-flyer is
// legal.
func TestDeclareBlockerFlyingCanBlockFlying(t *testing.T) {
	g := newActiveGame(t)
	atk := pushKeywordCreature(t, g, g.Seats[0], 2, 2, "flying")
	blk := pushKeywordCreature(t, g, g.Seats[1], 2, 2, "flying")
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(atk, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blk, atk); err != nil {
		t.Errorf("flyer vs flyer: %v", err)
	}
}

// TestCombatFirstStrikeBlockerDamagesVanillaAttacker verifies CR 510.2:
// a first-strike blocker deals damage in the first-strike substep even
// when the attacker itself has no first strike or double strike.
// Regression for the post-S18 bug where assignAndDealCombatDamageLocked
// skipped the entire attacker iteration (including the blocker-damage
// loop) when the attacker didn't participate in the substep.
func TestCombatFirstStrikeBlockerDamagesVanillaAttacker(t *testing.T) {
	g := newActiveGame(t)
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	knight := pushKeywordCreature(t, g, g.Seats[1], 2, 1, "first strike")
	ace := pushKeywordCreature(t, g, g.Seats[1], 1, 1, "double strike")
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.DeclareAttacker(bear, g.Seats[1].ID)
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.DeclareBlocker(knight, bear)
	g.DeclareBlocker(ace, bear)
	advanceIntoStep(t, g, StepCombatDamage)
	// First-strike substep: Knight (FS) deals 2, Ace (DS) deals 1 →
	// bear takes 3, dies via SBA before regular substep. Bear's 2
	// damage never lands (vanilla, no FS/DS).
	if findCard(g, bear) != nil {
		t.Errorf("bear should be dead from first-strike damage (2+1=3)")
	}
	// Knight and Ace both survive — bear was dead before regular.
	if findCard(g, knight) == nil {
		t.Errorf("knight should survive; bear died before regular substep")
	}
	if findCard(g, ace) == nil {
		t.Errorf("ace should survive; bear died before regular substep")
	}
}

// TestCombatUnblockedCommanderDamageEliminatesAt21 closes the loop
// the reviewer flagged: real combat never called
// RecordCommanderDamage, so the CR 903.10a 21-damage SBA could only
// fire from the manual set_commander_damage action. An unblocked
// 7-power commander connecting three times must now eliminate the
// defender through the production combat pipeline.
func TestCombatUnblockedCommanderDamageEliminatesAt21(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]

	cmdr := NewCommander("Test Commander", atk.ID)
	cmdr.TypeLine = "Legendary Creature — Avatar"
	cmdr.Power = 7
	cmdr.Toughness = 7
	g.Battlefield.PushTop(cmdr)

	for hit := 1; hit <= 3; hit++ {
		// Walk to seat 0's declare-attackers (skipping seat 1's
		// combat steps on intervening turns).
		for !(g.Turn.Step == StepDeclareAttackers && g.Turn.ActiveSeat == 0) {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("hit %d: AdvanceStep: %v", hit, err)
			}
		}
		if err := g.DeclareAttacker(cmdr.InstanceID, def.ID); err != nil {
			t.Fatalf("hit %d: DeclareAttacker: %v", hit, err)
		}
		advanceIntoStep(t, g, StepCombatDamage)
		if got, want := def.CommanderDamage[cmdr.InstanceID], 7*hit; got != want {
			t.Fatalf("hit %d: commander damage = %d, want %d", hit, got, want)
		}
	}

	if def.Life != StartingLife-21 {
		t.Errorf("defender life = %d, want %d", def.Life, StartingLife-21)
	}
	if !def.IsDeadByCommanderDamage() {
		t.Error("IsDeadByCommanderDamage = false after 21 commander damage")
	}
	if !def.Eliminated {
		t.Error("defender not eliminated by the 21-commander-damage SBA")
	}
	if g.State != StateEnded {
		t.Errorf("game state = %q, want %q (last survivor)", g.State, StateEnded)
	}
}

// TestCombatNonCommanderDamageNotRecorded guards the negative: a
// vanilla unblocked attacker must not pollute the commander-damage
// map.
func TestCombatNonCommanderDamageNotRecorded(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	bear := pushKeywordCreature(t, g, atk, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, def.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if def.Life != StartingLife-2 {
		t.Errorf("defender life = %d, want %d", def.Life, StartingLife-2)
	}
	if got := def.CommanderDamage[bear]; got != 0 {
		t.Errorf("non-commander attack recorded %d commander damage, want 0", got)
	}
}

// TestCombatCommanderTrampleOverflowRecordsCommanderDamage drives
// the markCombatDamageToPlayerFromFrameLocked path: a multi-blocked
// trampling commander queues the CR 510.1c assignment prompt, and
// the resume's trample-to-player spill must accrue commander damage
// from the frame's cached flags.
func TestCombatCommanderTrampleOverflowRecordsCommanderDamage(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]

	cmdr := NewCommander("Trampling Commander", atk.ID)
	cmdr.TypeLine = "Legendary Creature — Wurm"
	cmdr.Power = 6
	cmdr.Toughness = 6
	g.Battlefield.PushTop(cmdr)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cmdr.InstanceID {
			g.Battlefield.Cards[i].effective = printedEffectiveWith(cmdr, "trample")
		}
	}
	bear1 := pushKeywordCreature(t, g, def, 2, 2)
	bear2 := pushKeywordCreature(t, g, def, 3, 3)

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(cmdr.InstanceID, def.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(bear1, cmdr.InstanceID); err != nil {
		t.Fatalf("DeclareBlocker bear1: %v", err)
	}
	if err := g.DeclareBlocker(bear2, cmdr.InstanceID); err != nil {
		t.Fatalf("DeclareBlocker bear2: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)

	var promptID uuid.UUID
	for _, pc := range g.PendingChoices {
		if pc.Kind == PendingChoiceDamageAssignment {
			promptID = pc.ID
			break
		}
	}
	if promptID == uuid.Nil {
		t.Fatalf("expected damage_assignment prompt")
	}
	// 2 + 3 lethal across the bears, 1 trample to the player.
	if err := g.ResolveDamageAssignment(promptID, atk.ID,
		[]DamageAssignmentEntry{
			{BlockerID: bear1, Amount: 2},
			{BlockerID: bear2, Amount: 3},
		}, 1); err != nil {
		t.Fatalf("ResolveDamageAssignment: %v", err)
	}

	if def.Life != StartingLife-1 {
		t.Errorf("defender life = %d, want %d", def.Life, StartingLife-1)
	}
	if got := def.CommanderDamage[cmdr.InstanceID]; got != 1 {
		t.Errorf("trample overflow commander damage = %d, want 1", got)
	}
}

// TestCombatDamageEventCarriesCombatFlag — S19 sub-PR 7: the
// EventDealDamage emitted by the combat damage step is flagged
// Combat so "deals combat damage to a player" triggers can tell it
// apart from Lightning Bolt. Covers the unblocked-attacker path
// (player target) and the blocked path (creature target).
func TestCombatDamageEventCarriesCombatFlag(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	blockedAttacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	blocker := pushKeywordCreature(t, g, g.Seats[1], 4, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, id := range []uuid.UUID{attacker, blockedAttacker} {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, blockedAttacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)

	var toPlayer, toCreature, fromBlocker bool
	for _, ev := range g.Events {
		if ev.Kind != EventDealDamage || !ev.Combat {
			continue
		}
		switch {
		case ev.Source == attacker && ev.Target == g.Seats[1].ID && ev.Amount == 3:
			toPlayer = true
		case ev.Source == blockedAttacker && ev.Target == blocker:
			toCreature = true
		case ev.Source == blocker && ev.Target == blockedAttacker:
			fromBlocker = true
		}
	}
	if !toPlayer || !toCreature || !fromBlocker {
		t.Errorf("combat-flagged damage events: toPlayer=%v toCreature=%v fromBlocker=%v (want all true)",
			toPlayer, toCreature, fromBlocker)
	}

	// Non-combat damage stays unflagged.
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(attacker, g.Seats[1].ID, 1)
	})
	last := g.Events[len(g.Events)-1]
	if last.Kind != EventDealDamage || last.Combat {
		t.Errorf("effect damage event: kind=%s combat=%v, want deal_damage / false", last.Kind, last.Combat)
	}
}
