package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_cause_cards_test.go — #4534 and #4576, the two "becomes tapped"
// seams of Marvel Super Heroes (#4455): the cause a teamwork payment
// stamps on its tap (Agent Maria Hill), and the per-object
// became-tapped tally (Captain America, Living Legend). Engine-driven:
// real casts, real attack declarations, real triggers on the stack.

const (
	agentMariaHillOracle   = "a2d90efe-80f2-4069-b135-a8bd00ccd2b0"
	captainAmericaLLOracle = "67aa946d-e694-494e-ba0c-44880d5bc03e"
)

func pushMariaHill(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushDiesCreatureForTest(g, owner, "Agent Maria Hill", agentMariaHillOracle,
		"Legendary Creature — Human Spy Hero", 2, 1)
}

func pushCaptainAmerica(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushDiesCreatureForTest(g, owner, "Captain America, Living Legend", captainAmericaLLOracle,
		"Legendary Creature — Human Soldier Hero", 3, 4)
}

// tapPriority taps (or untaps) a permanent the way the sandbox does and
// then runs the priority boundary (CR 117.5), where the triggers the tap
// caused are put on the stack.
func tapPriority(g *game.Game, id uuid.UUID, tapped bool) error {
	err := g.TapCard(id, tapped)
	g.RunStateChecksForTest()
	return err
}

func tapCount(g *game.Game, id uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.TimesBecameTappedThisTurn(id) })
	return n
}

// --- Agent Maria Hill (#4534) -----------------------------------------

// castRepulsorBlastTeaming casts Repulsor Blast paying its teamwork
// cost with `team` and returns the caster's hand size once the cast is
// complete, so the draw can be told from the spell leaving the hand.
func castRepulsorBlastTeaming(t *testing.T, g *game.Game, me, victim uuid.UUID, team ...uuid.UUID) int {
	t.Helper()
	if _, err := castPaying(t, g, "Repulsor Blast", "Sorcery", repulsorBlastOracle, twTarget(victim),
		[]int{0}, team, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	return g.Seats[g.Turn.ActiveSeat].Hand.Size()
}

func TestAgentMariaHillTriggersWhenTappedToPayTeamwork(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	maria := pushMariaHill(g, me.ID)
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)

	handAfterCast := castRepulsorBlastTeaming(t, g, me.ID, victim, maria)
	if !twCard(g, maria).Tapped {
		t.Fatal("Maria Hill was not tapped to pay teamwork")
	}
	if triggerOnStack(g, maria) == nil {
		t.Fatal("Maria Hill's trigger is not on the stack after she was tapped to pay a teamwork cost")
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, maria); got != 1 {
		t.Errorf("+1/+1 counters on Maria Hill: %d, want 1", got)
	}
	if got := me.Hand.Size(); got != handAfterCast+1 {
		t.Errorf("hand size: %d, want %d (one card drawn)", got, handAfterCast+1)
	}
}

// TestAgentMariaHillIgnoresOtherTeamworkTaps: another creature tapped
// for the same payment is not Maria Hill becoming tapped.
func TestAgentMariaHillIgnoresOtherTeamworkTaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	maria := pushMariaHill(g, me.ID)
	other := pushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	victim := pushVanillaCreature(g, opp.ID, "Wall", 0, 10)

	castRepulsorBlastTeaming(t, g, me.ID, victim, other)
	if triggerOnStack(g, maria) != nil {
		t.Error("Maria Hill triggered on another creature paying the teamwork cost")
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, maria); got != 0 {
		t.Errorf("counters on an untapped Maria Hill: %d, want 0", got)
	}
}

// TestAgentMariaHillIgnoresOrdinaryTapsAndAttacks: becoming tapped for
// any other reason (a tap effect, a mana-style cost, attacking) is not
// paying a teamwork cost.
func TestAgentMariaHillIgnoresOrdinaryTapsAndAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	maria := pushMariaHill(g, me.ID)
	handBefore := me.Hand.Size()

	if err := tapPriority(g, maria, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if triggerOnStack(g, maria) != nil {
		t.Fatal("Maria Hill triggered on an ordinary tap")
	}
	if err := tapPriority(g, maria, false); err != nil {
		t.Fatalf("untap: %v", err)
	}

	attackWith(t, g, opp.ID, maria)
	passPriorityAroundTable(t, g)
	if !twCard(g, maria).Tapped {
		t.Fatal("Maria Hill did not tap to attack")
	}
	if got := plusOneCounters(g, maria); got != 0 {
		t.Errorf("counters after an ordinary tap and an attack: %d, want 0", got)
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand size %d, want %d — nothing should have been drawn", got, handBefore)
	}
}

// TestAgentMariaHillTriggerSurvivesARestore: the tap cause rides the
// trigger's context, so a table restored with the trigger still on the
// stack resolves it.
func TestAgentMariaHillTriggerSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	maria := pushMariaHill(g, me.ID)
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)
	handAfterCast := castRepulsorBlastTeaming(t, g, me.ID, victim, maria)
	if triggerOnStack(g, maria) == nil {
		t.Fatal("no trigger to restore")
	}

	restored := restoreRoundTrip(t, g, false)
	if triggerOnStack(restored, maria) == nil {
		t.Fatal("the restore dropped Maria Hill's trigger")
	}
	passPriorityAroundTable(t, restored)
	if got := plusOneCounters(restored, maria); got != 1 {
		t.Errorf("after a restore, counters on Maria Hill: %d, want 1", got)
	}
	var seat *game.Player
	for _, s := range restored.Seats {
		if s.ID == me.ID {
			seat = s
		}
	}
	if got := seat.Hand.Size(); got != handAfterCast+1 {
		t.Errorf("after a restore, hand size %d, want %d", got, handAfterCast+1)
	}
}

// --- Captain America, Living Legend (#4576) ----------------------------

// captainBoard is the common table: Captain America and a vanilla Bear
// under the active player, in their precombat main phase.
func captainBoard(t *testing.T) (g *game.Game, me, opp *game.Player, captain, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	captain = pushCaptainAmerica(g, me.ID)
	bear = pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	return g, me, opp, captain, bear
}

func TestCaptainAmericaUntapsAFirstTapDuringYourTurn(t *testing.T) {
	g, _, _, captain, bear := captainBoard(t)

	if err := tapPriority(g, bear, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if triggerOnStack(g, captain) == nil {
		t.Fatal("the first tap of a creature you control on your turn did not trigger")
	}
	passPriorityAroundTable(t, g)
	if twCard(g, bear).Tapped {
		t.Error("the Bear is still tapped after its first tap of the turn")
	}

	// The second tap is not the first time.
	if err := tapPriority(g, bear, true); err != nil {
		t.Fatalf("second TapCard: %v", err)
	}
	if triggerOnStack(g, captain) != nil {
		t.Error("the second tap of the same creature triggered")
	}
	passPriorityAroundTable(t, g)
	if !twCard(g, bear).Tapped {
		t.Error("the Bear was untapped on its second tap")
	}
	if got := tapCount(g, bear); got != 2 {
		t.Errorf("became-tapped tally for the Bear: %d, want 2", got)
	}
}

// TestCaptainAmericaCountsEachCreatureOnItsOwn: the first time is a
// fact about each creature, and Captain America's own taps count too.
func TestCaptainAmericaCountsEachCreatureOnItsOwn(t *testing.T) {
	g, me, _, captain, bear := captainBoard(t)
	second := pushVanillaCreature(g, me.ID, "Second Bear", 2, 2)

	for _, id := range []uuid.UUID{bear, second, captain} {
		if err := tapPriority(g, id, true); err != nil {
			t.Fatalf("TapCard: %v", err)
		}
		passPriorityAroundTable(t, g)
		if twCard(g, id).Tapped {
			t.Errorf("%s was still tapped after its first tap of the turn", twCard(g, id).Name)
		}
	}
}

func TestCaptainAmericaDoesNothingOnAnOpponentsTurn(t *testing.T) {
	g, _, opp, captain, bear := captainBoard(t)
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if g.Seats[g.Turn.ActiveSeat].ID != opp.ID {
		t.Fatalf("expected the opponent's turn")
	}

	if err := tapPriority(g, bear, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if triggerOnStack(g, captain) != nil {
		t.Error("a creature tapped on an opponent's turn triggered")
	}
	passPriorityAroundTable(t, g)
	if !twCard(g, bear).Tapped {
		t.Error("the Bear was untapped on an opponent's turn")
	}
}

func TestCaptainAmericaIgnoresAnOpponentsCreatures(t *testing.T) {
	g, _, opp, captain, _ := captainBoard(t)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	if err := tapPriority(g, theirs, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if triggerOnStack(g, captain) != nil {
		t.Error("an opponent's creature tapped on my turn triggered")
	}
	passPriorityAroundTable(t, g)
	if !twCard(g, theirs).Tapped {
		t.Error("an opponent's creature was untapped")
	}
}

// TestCaptainAmericaUntapsAnAttackerButNotAVigilantOne: declaring an
// attacker taps it (CR 508.1f) without an EventTapCard, and that is its
// first tap. A vigilance attacker does not tap, so nothing triggers and
// nothing is counted.
func TestCaptainAmericaUntapsAnAttackerButNotAVigilantOne(t *testing.T) {
	g, _, opp, captain, bear := captainBoard(t)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{captain, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if !twCard(g, bear).Tapped {
		t.Fatal("the Bear did not tap to attack")
	}
	lockInAttacks(t, g)
	if triggerOnStack(g, captain) == nil {
		t.Fatal("the Bear's attack tap did not trigger Captain America")
	}
	passPriorityAroundTable(t, g)

	if twCard(g, bear).Tapped {
		t.Error("the attacking Bear was not untapped")
	}
	if twCard(g, bear).AttackingTarget == uuid.Nil {
		t.Error("untapping the Bear took it out of combat")
	}
	if got := tapCount(g, bear); got != 1 {
		t.Errorf("became-tapped tally for the attacking Bear: %d, want 1", got)
	}
	if twCard(g, captain).Tapped {
		t.Error("Captain America (vigilance) tapped to attack")
	}
	if got := tapCount(g, captain); got != 0 {
		t.Errorf("became-tapped tally for a vigilance attacker: %d, want 0", got)
	}
}

// TestCaptainAmericaAttackTapIsNotTheFirstAfterAnEarlierTap: a creature
// tapped for something earlier this turn, then attacking, is tapped for
// the second time and stays tapped.
func TestCaptainAmericaAttackTapIsNotTheFirstAfterAnEarlierTap(t *testing.T) {
	g, _, opp, captain, bear := captainBoard(t)

	if err := tapPriority(g, bear, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if twCard(g, bear).Tapped {
		t.Fatal("the first tap was not undone")
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if triggerOnStack(g, captain) != nil {
		t.Error("the attack tap was the Bear's second tap, but it triggered")
	}
	passPriorityAroundTable(t, g)
	if !twCard(g, bear).Tapped {
		t.Error("the Bear was untapped by its second tap")
	}
}

// TestCaptainAmericaTallyResetsNextTurn: the next turn's first tap is a
// first tap again.
func TestCaptainAmericaTallyResetsNextTurn(t *testing.T) {
	g, me, _, captain, bear := captainBoard(t)
	seat := g.Turn.ActiveSeat

	if err := tapPriority(g, bear, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := tapPriority(g, bear, true); err != nil {
		t.Fatalf("second TapCard: %v", err)
	}
	if got := tapCount(g, bear); got != 2 {
		t.Fatalf("tally this turn: %d, want 2", got)
	}

	advanceToMainOf(t, g, (seat+1)%4)
	advanceToMainOf(t, g, seat)
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatal("not back on my turn")
	}
	if got := tapCount(g, bear); got != 0 {
		t.Errorf("the tally survived the turn: %d, want 0", got)
	}
	if err := tapPriority(g, bear, true); err != nil {
		t.Fatalf("TapCard next turn: %v", err)
	}
	if triggerOnStack(g, captain) == nil {
		t.Fatal("the first tap of the next turn did not trigger")
	}
	passPriorityAroundTable(t, g)
	if twCard(g, bear).Tapped {
		t.Error("the Bear was not untapped by its first tap of the next turn")
	}
}

// TestCaptainAmericaTreatsAFlickeredCreatureAsNew: a creature that left
// and came back is a new object (CR 400.7), so its next tap is its
// first. Two returns: one that keeps the instance ID (a bounce, so only
// the object epoch tells the objects apart) and one that mints a new
// one (an exile return).
func TestCaptainAmericaTreatsAFlickeredCreatureAsNew(t *testing.T) {
	for _, tc := range []struct {
		name string
		back func(t *testing.T, g *game.Game, id, owner uuid.UUID) uuid.UUID
	}{
		{"bounced", func(t *testing.T, g *game.Game, id, owner uuid.UUID) uuid.UUID {
			hand := game.ZoneRef{Kind: game.ZoneHand, Owner: owner}
			field := game.ZoneRef{Kind: game.ZoneBattlefield}
			if err := g.MoveCardByID(field, hand, id); err != nil {
				t.Fatalf("bounce: %v", err)
			}
			if err := g.MoveCardByID(hand, field, id); err != nil {
				t.Fatalf("return: %v", err)
			}
			return id
		}},
		{"exiled and returned", func(t *testing.T, g *game.Game, id, owner uuid.UUID) uuid.UUID {
			var back uuid.UUID
			g.WithWriteLock(func() {
				if err := g.ExileCardForEffect(id); err != nil {
					t.Fatalf("exile: %v", err)
				}
				var err error
				if back, err = g.ReturnFromExileToBattlefieldForEffect(id, owner, false); err != nil {
					t.Fatalf("return: %v", err)
				}
			})
			return back
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, captain, bear := captainBoard(t)

			if err := tapPriority(g, bear, true); err != nil {
				t.Fatalf("TapCard: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := tapCount(g, bear); got != 1 {
				t.Fatalf("tally before the flicker: %d, want 1", got)
			}

			back := tc.back(t, g, bear, me.ID)
			if findBattlefieldCardByID(g, back) == nil {
				t.Fatal("the creature did not come back")
			}
			if got := tapCount(g, back); got != 0 {
				t.Errorf("tally for the returned creature: %d, want 0 (a new object)", got)
			}
			if err := tapPriority(g, back, true); err != nil {
				t.Fatalf("TapCard after the flicker: %v", err)
			}
			if triggerOnStack(g, captain) == nil {
				t.Fatal("the new object's first tap did not trigger")
			}
			passPriorityAroundTable(t, g)
			if twCard(g, back).Tapped {
				t.Error("the returned creature was not untapped on its first tap")
			}
		})
	}
}

// TestCaptainAmericaTallySurvivesARestore: a restored table remembers
// that the creature already became tapped this turn, and also resolves a
// trigger that was waiting on the stack.
func TestCaptainAmericaTallySurvivesARestore(t *testing.T) {
	g, _, _, captain, bear := captainBoard(t)
	second := pushVanillaCreature(g, g.Seats[g.Turn.ActiveSeat].ID, "Second Bear", 2, 2)

	if err := tapPriority(g, bear, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The second Bear's trigger is left on the stack across the restore.
	if err := tapPriority(g, second, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if triggerOnStack(g, captain) == nil {
		t.Fatal("no trigger to restore")
	}

	restored := restoreRoundTrip(t, g, false)
	if got := tapCount(restored, bear); got != 1 {
		t.Fatalf("restored tally for the first Bear: %d, want 1", got)
	}
	passPriorityAroundTable(t, restored)
	if twCard(restored, second).Tapped {
		t.Error("the restored trigger did not untap the second Bear")
	}

	if err := tapPriority(restored, bear, true); err != nil {
		t.Fatalf("TapCard after the restore: %v", err)
	}
	if triggerOnStack(restored, captain) != nil {
		t.Error("a second tap after a restore was taken for the first")
	}
	passPriorityAroundTable(t, restored)
	if !twCard(restored, bear).Tapped {
		t.Error("the Bear was untapped by its second tap after a restore")
	}
}
