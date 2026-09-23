package game

import (
	"testing"

	"github.com/google/uuid"
)

// attacking_entry_test.go — #1227, the two engine halves ninjutsu
// (CR 702.49) needed and nothing in the catalog could ask for:
//
//  1. Game.UnblockedAttackerForEffect — CR 509.1h read from outside
//     internal/game for the first time, so a cost clause can say
//     "an unblocked attacker you control".
//  2. ZoneEntryOptions.Attacking — an entry that puts a CARD onto the
//     battlefield attacking (CR 506.3c). Before it only a minted token
//     could enter attacking.
//
// Plus the fact that crosses between them: PaidCost.ReturnedAttacking,
// what the permanent a return-to-hand cost returned was attacking,
// read by the payer BEFORE the bounce because the exit clears it.
//
// The card-level keyword is cards/effects/ninjutsu_test.go.

// --- 1. the unblocked-attacker read -------------------------------

// The declare-ATTACKERS step is too early. Nothing is blocked OR
// unblocked until blockers are declared (CR 509.1h decides it as part
// of that declaration), and the blocked record is empty until the
// lock-in writes it — so without the step clause every attacker would
// read as unblocked here and ninjutsu would be a main-phase ability.
func TestUnblockedAttackerNeedsTheDeclareBlockersStep(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	g.ReadSnapshot(func() {
		if g.UnblockedAttackerForEffect(attacker) {
			t.Error("an attacker in the declare-attackers step reads as unblocked — blockers have not been declared yet (CR 509.1h)")
		}
	})
	advanceIntoStep(t, g, StepDeclareBlockers)
	g.ReadSnapshot(func() {
		if !g.UnblockedAttackerForEffect(attacker) {
			t.Error("an unblocked attacker in the declare-blockers step does not read as unblocked")
		}
	})
}

// A creature that is not attacking at all is never an unblocked
// attacker, whatever the step.
func TestUnblockedAttackerIsFalseForACreatureNotAttacking(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	bystander := pushCombatant(t, g, g.Seats[0], "Bystander", 1, 1)
	declareAttacks(t, g, attacker)

	g.ReadSnapshot(func() {
		if g.UnblockedAttackerForEffect(bystander) {
			t.Error("a creature that never attacked reads as an unblocked attacker")
		}
	})
}

// The blocked RECORD, which is the fact that outlives the blocker
// (CR 510.1c, #715): once the lock-in has run, the attacker is blocked
// and stays blocked even after its blocker dies.
func TestUnblockedAttackerIsFalseForABlockedAttacker(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	chump := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
	blockAfterLockIn(t, g, attacker, chump)

	g.ReadSnapshot(func() {
		if g.UnblockedAttackerForEffect(attacker) {
			t.Error("a blocked attacker reads as unblocked")
		}
	})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(chump); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	g.ReadSnapshot(func() {
		if g.UnblockedAttackerForEffect(attacker) {
			t.Error("a blocked attacker whose only blocker died reads as unblocked — CR 509.1h keeps it blocked")
		}
	})
}

// And the half the record cannot answer: DeclareBlocker only STAGES
// the pairing (#830), so between the defender's click and the next
// priority boundary the blocked record is still empty. An attacker
// with a blocker already pointed at it must not read as unblocked in
// that window, or ninjutsu would beat a block that has been made.
func TestUnblockedAttackerIsFalseForAStagedBlockNotYetLockedIn(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	chump := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
	declareAttacks(t, g, attacker)
	if err := g.DeclareBlockers([]BlockDeclaration{{Blocker: chump, Attacker: attacker}}); err != nil {
		t.Fatalf("DeclareBlockers: %v", err)
	}
	if g.blockedAttackers[attacker] {
		t.Fatalf("setup: the declaration locked itself in; there is no staged window to test")
	}

	g.ReadSnapshot(func() {
		if g.UnblockedAttackerForEffect(attacker) {
			t.Error("an attacker with a staged, unannounced blocker reads as unblocked")
		}
	})
}

// --- 2. the attacking entry ---------------------------------------

// pushHandCreature puts a creature card in a player's hand and returns
// its instance ID.
func pushHandCreature(t *testing.T, g *Game, owner *Player, name string, power, toughness int) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.Power = power
	c.Toughness = toughness
	owner.Hand.PushTop(c)
	return c.InstanceID
}

// attackEventsFor counts the EventAttack events naming `id`.
func attackEventsFor(g *Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventAttack && ev.CardID == id {
			n++
		}
	}
	return n
}

// The headline: a card put onto the battlefield from hand with
// ZoneEntryOptions.Attacking arrives tapped and attacking, and
// CR 506.3c holds — it was never DECLARED, so the attack-declaration
// lock-in must not announce it and no "whenever ~ attacks" trigger
// can see it.
func TestPutFromHandOntoBattlefieldAttackingIsNeverDeclared(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	ninja := pushHandCreature(t, g, me, "Ninja", 2, 2)
	declareAttacks(t, g, attacker)

	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.PutFromHandOntoBattlefieldForEffect(ninja, HandEntryOptions{
			Controller: me.ID,
			Tapped:     true,
			Attacking:  opp.ID,
		})
	})
	if err != nil {
		t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
	}
	if entered != ninja {
		t.Fatalf("entered %s, want %s", entered, ninja)
	}
	c := findBattlefieldCard(g, ninja)
	if c == nil {
		t.Fatal("the card did not reach the battlefield")
	}
	if !c.Tapped {
		t.Error("the card entered untapped")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("AttackingTarget = %s, want %s", c.AttackingTarget, opp.ID)
	}
	if !g.announcedAttacks[ninja] {
		t.Error("the entry did not mark the permanent announced — the lock-in would hand it an attack trigger (CR 506.3c)")
	}
	g.WithWriteLock(func() { g.commitAttackDeclarationLocked() })
	if n := attackEventsFor(g, ninja); n != 0 {
		t.Errorf("%d EventAttack for a permanent PUT onto the battlefield attacking, want 0 (CR 506.3c)", n)
	}
	if n := attackEventsFor(g, attacker); n != 1 {
		t.Errorf("%d EventAttack for the creature that WAS declared, want 1", n)
	}
}

// A defender that is no longer attackable leaves the permanent on the
// battlefield not attacking rather than erroring — the posture
// CreateTokensAttackingForEffect already took, said once for both
// entry doors.
func TestPutOntoBattlefieldAttackingIgnoresAnUnattackableDefender(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ninja := pushHandCreature(t, g, me, "Ninja", 2, 2)
	advanceIntoStep(t, g, StepDeclareBlockers)

	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(ninja, HandEntryOptions{
			Controller: me.ID,
			Tapped:     true,
			// A uuid that names nothing — a planeswalker that died
			// while the ability was on the stack.
			Attacking: uuid.New(),
		}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	c := findBattlefieldCard(g, ninja)
	if c == nil {
		t.Fatal("the card did not reach the battlefield")
	}
	if c.AttackingTarget != uuid.Nil {
		t.Errorf("AttackingTarget = %s, want none — the defender is not attackable", c.AttackingTarget)
	}
	if g.announcedAttacks[ninja] {
		t.Error("a permanent that entered NOT attacking was marked announced")
	}
}

// The point of the whole feature: a creature that enters attacking
// after blockers have been declared is unblocked, so it connects in
// the combat damage step.
func TestCreatureThatEnteredAttackingDealsUnblockedCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	ninja := pushHandCreature(t, g, me, "Ninja", 3, 3)
	chump := pushCombatant(t, g, opp, "Chump", 1, 1)
	blockAfterLockIn(t, g, attacker, chump)

	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(ninja, HandEntryOptions{
			Controller: me.ID,
			Tapped:     true,
			Attacking:  opp.ID,
		}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	before := opp.Life
	passUntilStep(t, g, StepCombatDamage)
	passUntilStep(t, g, StepEndCombat)

	if got := before - opp.Life; got != 3 {
		t.Errorf("defender lost %d life, want 3 — the entering attacker is unblocked and the Bears are chumped", got)
	}
}

// --- 3. the paid-cost record --------------------------------------

// returnCreatureCost is "return a creature you control to its owner's
// hand" — the in-package stand-in for ninjutsu's clause, without the
// unblocked narrowing, so the test pins the RECORD rather than the
// filter.
func returnCreatureCost() AbilityCost {
	return AbilityCost{ReturnToHand: &ReturnToHandCost{
		Count: 1,
		Filter: &TargetSpec{
			Mode:  "permanent",
			Label: "a creature you control",
			Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 1,
		},
		Label: "a creature you control",
	}}
}

// PaidCost.ReturnedAttacking: read BEFORE the bounce, because the exit
// clears Card.AttackingTarget and LKI carries no combat state. Without
// it ninjutsu could not know which player its ninja is attacking.
func TestReturnCostRecordsWhatTheReturnedAttackerWasAttacking(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	src := pushReturnCostSource(g, me, returnCreatureCost())
	declareAttacks(t, g, attacker)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{attacker}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Fatal("the attacker was not returned to hand")
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want 1", len(g.StackMeta))
	}
	for _, item := range g.StackMeta {
		if item.Paid.ReturnedAttacking != opp.ID {
			t.Errorf("PaidCost.ReturnedAttacking = %s, want %s — the attack has to be read before the bounce clears it",
				item.Paid.ReturnedAttacking, opp.ID)
		}
	}
}

// A return cost that returns something which was not attacking records
// nothing, which is every printed return cost but ninjutsu's.
func TestReturnCostOfANonAttackerRecordsNoAttack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bystander := pushCombatant(t, g, me, "Bystander", 1, 1)
	src := pushReturnCostSource(g, me, returnCreatureCost())
	advanceTo(t, g, StepPrecombatMain)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{bystander}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, item := range g.StackMeta {
		if item.Paid.ReturnedAttacking != uuid.Nil {
			t.Errorf("PaidCost.ReturnedAttacking = %s for a permanent that was attacking nothing, want none",
				item.Paid.ReturnedAttacking)
		}
	}
}

// The token door's version of the same case, and the reason the stamp
// CLEARS rather than bailing out quietly: a token is minted with its
// AttackingTarget already on it, so a defender that stopped being
// attackable between the minting and the entry would otherwise leave
// the token attacking nobody AND unmarked — and the next lock-in would
// read that as a staged declaration and hand it an attack trigger.
func TestEntryAttackerStampClearsAStaleAttack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	tok := pushCombatant(t, g, me, "Angel Token", 4, 4)
	g.WithWriteLock(func() {
		findBattlefieldCard(g, tok).AttackingTarget = uuid.New() // a seat that has left
		g.stampEntryAttackerLocked(tok, uuid.New())
	})

	if c := findBattlefieldCard(g, tok); c.AttackingTarget != uuid.Nil {
		t.Errorf("AttackingTarget = %s, want none — the defender is not attackable", c.AttackingTarget)
	}
	if g.announcedAttacks[tok] {
		t.Error("a permanent that is not attacking was marked announced")
	}
	g.WithWriteLock(func() { g.commitAttackDeclarationLocked() })
	if n := attackEventsFor(g, tok); n != 0 {
		t.Errorf("%d EventAttack for a permanent attacking nothing, want 0", n)
	}
}
