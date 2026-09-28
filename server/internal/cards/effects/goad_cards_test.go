package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// goad_cards_test.go — #1599's "up to six more" one-file goad cards:
// Jeering Homunculus, Taunting Kobold, Vengeful Ancestor, Goblin
// Racketeer, Coveted Peacock, Puppet Master, String Puller. All built
// on GoadTarget (goad.go) or TargetCreatureDefendingPlayerControls
// (defending_player_targets.go).

const (
	jeeringHomunculusOracle = "146de263-fece-4a00-ab67-35dc934165d8"
	tauntingKobaldOracle    = "604c581e-c1ba-417c-b9de-182192cc85cb"
	vengefulAncestorOracle  = "f3b3173b-f7ae-420e-84ec-ea61414674a0"
	goblinRacketeerOracle   = "9336a62c-f2f9-45a8-bf69-86060aa0ce59"
	covetedPeacockOracle    = "e6ad5e92-c1ab-4c91-95ca-1af295e71b23"
	puppetMasterOracle      = "376c6257-4a1a-42d4-8931-72068f38727f"
)

// --- Jeering Homunculus ---------------------------------------------

// TestJeeringHomunculusMayGoadTargetCreature: an ETB "you may" goad,
// and the refusal it produces once accepted.
func TestJeeringHomunculusMayGoadTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})

	castAndResolveCreature(t, g, "Jeering Homunculus", "Creature — Homunculus", jeeringHomunculusOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if c, ok := g.LookupCardForEffect(bear); !ok || c.GoadedBy != me.ID {
		t.Fatal("Jeering Homunculus did not goad the bear")
	}

	advanceToDeclareAttackersOf(t, g, oppSeat)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Fatal("attacking the goader succeeded, want the \"other than you\" half refused")
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with the goaded bear home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the goaded bear attacking someone else: %v", err)
	}
}

// --- Taunting Kobold --------------------------------------------------

// TestTauntingKobaldGoadsOnAttack: the attack trigger goads an
// opponent's creature, mandatorily, and the goad's refusal follows.
func TestTauntingKobaldGoadsOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	kobold := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Taunting Kobold", TypeLine: "Creature — Kobold",
		OracleID: tauntingKobaldOracle, Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})

	advanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(kobold, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Kobold): %v", err)
	}
	lockInAttacks(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if c, ok := g.LookupCardForEffect(bear); !ok || c.GoadedBy != me.ID {
		t.Fatal("Taunting Kobold did not goad the bear")
	}

	advanceToDeclareAttackersOf(t, g, oppSeat)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Fatal("attacking the goader succeeded, want the \"other than you\" half refused")
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with the goaded bear home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the goaded bear attacking someone else: %v", err)
	}
}

// --- Vengeful Ancestor ------------------------------------------------

// TestVengefulAncestorGoadsOnETBAndDamagesAnyGoadedAttacker: the ETB
// half goads mandatorily (no "you may"); the second ability then
// deals 1 damage to ANY goaded attacker's controller — not only one
// this card goaded — when it attacks.
func TestVengefulAncestorGoadsOnETBAndDamagesAnyGoadedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})

	castAndResolveCreature(t, g, "Vengeful Ancestor", "Creature — Spirit Dragon", vengefulAncestorOracle)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if c, ok := g.LookupCardForEffect(bear); !ok || c.GoadedBy != me.ID {
		t.Fatal("Vengeful Ancestor's ETB did not goad the bear")
	}

	advanceToDeclareAttackersOf(t, g, oppSeat)
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with the goaded bear home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	oppLifeBefore := opp.Life
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if opp.Life != oppLifeBefore-1 {
		t.Errorf("opponent's life = %d, want %d (1 damage from the goaded attack trigger)", opp.Life, oppLifeBefore-1)
	}
}

// --- Goblin Racketeer / Coveted Peacock --------------------------------

// racketeerShapeTest is the shared body for Goblin Racketeer and
// Coveted Peacock: identical printed ability, same test.
func racketeerShapeTest(t *testing.T, name, typeLine, oracleID string) {
	t.Helper()
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		OracleID: oracleID, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})

	advanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if c, ok := g.LookupCardForEffect(victim); !ok || c.GoadedBy != me.ID {
		t.Fatalf("%s did not goad the defending player's creature", name)
	}

	advanceToDeclareAttackersOf(t, g, oppSeat)
	if err := g.DeclareAttacker(victim, me.ID); err == nil {
		t.Fatal("attacking the goader succeeded, want the \"other than you\" half refused")
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with the goaded creature home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(victim, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the goaded creature attacking someone else: %v", err)
	}
}

func TestGoblinRacketeerMayGoadDefendingPlayersCreature(t *testing.T) {
	racketeerShapeTest(t, "Goblin Racketeer", "Creature — Goblin Rogue", goblinRacketeerOracle)
}

func TestCovetedPeacockMayGoadDefendingPlayersCreature(t *testing.T) {
	racketeerShapeTest(t, "Coveted Peacock", "Creature — Bird", covetedPeacockOracle)
}

// --- Puppet Master, String Puller --------------------------------------

// TestPuppetMasterGoadsAndPreventsBlockOnAttack: "whenever you
// attack" fires once per combat (not per attacker — OncePerBatch),
// goads an opponent's creature and stops it blocking this turn.
func TestPuppetMasterGoadsAndPreventsBlockOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	puppet := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Puppet Master, String Puller", TypeLine: "Legendary Creature — Human Artificer Villain",
		OracleID: puppetMasterOracle, Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	other := pushVanillaCreature(g, me.ID, "My Other Bear", 2, 2)
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})

	advanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(puppet, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Puppet Master): %v", err)
	}
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(other attacker): %v", err)
	}
	lockInAttacks(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(victim)
	if !ok || c.GoadedBy != me.ID {
		t.Fatal("Puppet Master did not goad the opponent's creature")
	}
	if !game.Restricted(&c, game.CantBlock) {
		t.Error("the goaded creature can still block this turn")
	}

	advanceToDeclareAttackersOf(t, g, oppSeat)
	if err := g.DeclareAttacker(victim, me.ID); err == nil {
		t.Fatal("attacking the goader succeeded, want the \"other than you\" half refused")
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with the goaded creature home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(victim, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the goaded creature attacking someone else: %v", err)
	}
}

// TestPuppetMasterTreasureFromAnyGoadedCreatureConnecting: the second
// ability reads GoadedBy alone — a creature goaded by someone else's
// effect still pays off.
func TestPuppetMasterTreasureFromAnyGoadedCreatureConnecting(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Puppet Master, String Puller", puppetMasterOracle, "Legendary Creature — Human Artificer Villain")
	goaded := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goaded Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Keywords: []string{"haste"},
		GoadedBy: me.ID,
	})

	attackWith(t, g, opp.ID, goaded)
	passPriorityAroundTable(t, g)

	treasures := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasures++
		}
	}
	if treasures != 1 {
		t.Errorf("treasures = %d, want 1", treasures)
	}
}
