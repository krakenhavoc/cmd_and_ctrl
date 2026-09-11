package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// attack_target_test.go — S27, CR 506.2 / 508.1d / 120.3, and #406.
//
// The two halves have to be tested together, because shipping either
// alone is worse than shipping neither: a declarable attack on a
// planeswalker that removes no loyalty is a silent no-op that looks
// like it worked, and a damage rule nothing can reach is untested
// code.

// pushPlaneswalkerForTest puts a planeswalker with `loyalty` counters
// on the battlefield under `owner`.
func pushPlaneswalkerForTest(g *Game, owner uuid.UUID, name string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Legendary Planeswalker — Test",
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{CounterLoyalty: loyalty},
	})
	return id
}

// pushBattleForTest puts a battle with `defense` counters and an
// explicit protector on the battlefield under `owner`.
func pushBattleForTest(g *Game, owner, protector uuid.UUID, name string, defense int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID:        id,
		Name:              name,
		TypeLine:          "Battle — Siege",
		Owner:             owner,
		Controller:        owner,
		ProtectorPlayerID: protector,
		Counters:          map[string]int{CounterDefense: defense},
	})
	return id
}

// pushReadyAttackerForTest puts an untapped, non-sick creature on the
// battlefield.
func pushReadyAttackerForTest(g *Game, owner uuid.UUID, name string, power int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		Power:      power,
		Toughness:  power,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func stepToDeclareAttackers(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if g.Turn.Step == StepDeclareAttackers {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached declare_attackers")
}

func counterOn(g *Game, id uuid.UUID, name string) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[name]
		}
	}
	return -1
}

// --- CR 120.3: what damage DOES ---------------------------------

// TestDamageRemovesLoyaltyFromAPlaneswalker is #406. Before the
// CR 120.3 split, a Lightning Bolt aimed at a planeswalker
// incremented DamageMarked — a number the lethal-damage SBA only
// reads inside its IsCreature() branch — so planeswalkers were
// unkillable by damage and the spell was a silent no-op.
func TestDamageRemovesLoyaltyFromAPlaneswalker(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	pw := pushPlaneswalkerForTest(g, owner, "Test Walker", 5)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, pw, 3); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})

	if got := counterOn(g, pw, CounterLoyalty); got != 2 {
		t.Errorf("loyalty after 3 damage to a 5-loyalty walker = %d, want 2", got)
	}
}

// TestLethalDamageToAPlaneswalkerKillsIt closes the loop: the loyalty
// reaches zero and the CR 704.5i state-based action takes it.
func TestLethalDamageToAPlaneswalkerKillsIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	pw := pushPlaneswalkerForTest(g, owner, "Test Walker", 3)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, pw, 4); err != nil {
			t.Fatalf("damage: %v", err)
		}
		g.runStateChecksLocked()
	})

	if counterOn(g, pw, CounterLoyalty) >= 0 {
		t.Error("a planeswalker dealt more than lethal damage survived")
	}
}

// TestDamageRemovesDefenseFromABattle is CR 120.3e, the same rule one
// card type over.
func TestDamageRemovesDefenseFromABattle(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, protector := g.Seats[0].ID, g.Seats[1].ID
	b := pushBattleForTest(g, owner, protector, "Test Siege", 5)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, b, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})

	if got := counterOn(g, b, CounterDefense); got != 3 {
		t.Errorf("defense after 2 damage to a 5-defense battle = %d, want 3", got)
	}
}

// TestDamageToAPlaneswalkerIgnoresCounterDoublers pins the reason the
// removal does NOT go through AddCounterForEffect. Removal is not
// addition: a Doubling Season must not double the loyalty coming OFF
// a walker, which is what routing through the CR 614 counter pipeline
// would do.
func TestDamageToAPlaneswalkerIgnoresCounterDoublers(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	pw := pushPlaneswalkerForTest(g, owner, "Test Walker", 6)

	// A synthetic doubler on the counter pipeline, standing in for
	// Doubling Season without needing the catalog.
	doubled := 0
	g.WithWriteLock(func() {
		g.testReplacements = append(g.testReplacements, ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				doubled++
				ev.CounterDelta *= 2
				return nil
			},
			Label: "test doubler",
		})
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, pw, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})

	if doubled != 0 {
		t.Errorf("the counter-replacement pipeline saw %d damage-driven removals; it must see none", doubled)
	}
	if got := counterOn(g, pw, CounterLoyalty); got != 4 {
		t.Errorf("loyalty after 2 damage = %d, want 4 (not doubled to 2)", got)
	}
}

// --- CR 506.2 / 508.1d: what you may attack ---------------------

func TestDeclareAttackerAcceptsAPlaneswalker(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := g.Seats[g.Turn.ActiveSeat].ID
	defender := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	pw := pushPlaneswalkerForTest(g, defender, "Their Walker", 4)
	bear := pushReadyAttackerForTest(g, attacker, "Bear", 2)
	stepToDeclareAttackers(t, g)

	if err := g.DeclareAttacker(bear, pw); err != nil {
		t.Fatalf("attacking a planeswalker: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == bear && c.AttackingTarget != pw {
			t.Errorf("AttackingTarget = %v, want the planeswalker", c.AttackingTarget)
		}
	}
}

func TestDeclareAttackerRejectsYourOwnPlaneswalker(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := g.Seats[g.Turn.ActiveSeat].ID
	pw := pushPlaneswalkerForTest(g, attacker, "My Walker", 4)
	bear := pushReadyAttackerForTest(g, attacker, "Bear", 2)
	stepToDeclareAttackers(t, g)

	if err := g.DeclareAttacker(bear, pw); !errors.Is(err, ErrIllegalAttackTarget) {
		t.Fatalf("attacking your own planeswalker: err = %v, want ErrIllegalAttackTarget", err)
	}
}

func TestDeclareAttackerRejectsABattleYouProtect(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := g.Seats[g.Turn.ActiveSeat].ID
	caster := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	// The active player is the protector, so the battle is theirs to
	// defend and not theirs to attack (CR 310.7).
	b := pushBattleForTest(g, caster, attacker, "Their Siege", 5)
	bear := pushReadyAttackerForTest(g, attacker, "Bear", 2)
	stepToDeclareAttackers(t, g)

	if err := g.DeclareAttacker(bear, b); !errors.Is(err, ErrIllegalAttackTarget) {
		t.Fatalf("attacking a battle you protect: err = %v, want ErrIllegalAttackTarget", err)
	}
}

func TestDeclareAttackerAcceptsABattleYouDoNotProtect(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := g.Seats[g.Turn.ActiveSeat].ID
	caster := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	protector := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID
	b := pushBattleForTest(g, caster, protector, "Their Siege", 5)
	bear := pushReadyAttackerForTest(g, attacker, "Bear", 2)
	stepToDeclareAttackers(t, g)

	if err := g.DeclareAttacker(bear, b); err != nil {
		t.Fatalf("attacking a battle someone else protects: %v", err)
	}
}

// TestCombatDamageToAPlaneswalkerRemovesLoyalty is the two halves
// meeting: declare, let the damage step resolve, and check the
// loyalty rather than a life total.
func TestCombatDamageToAPlaneswalkerRemovesLoyalty(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := g.Seats[g.Turn.ActiveSeat].ID
	defenderSeat := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lifeBefore := defenderSeat.Life
	pw := pushPlaneswalkerForTest(g, defenderSeat.ID, "Their Walker", 5)
	bear := pushReadyAttackerForTest(g, attacker, "Bear", 3)
	stepToDeclareAttackers(t, g)

	if err := g.DeclareAttacker(bear, pw); err != nil {
		t.Fatalf("declare: %v", err)
	}
	for i := 0; i < 10 && g.Turn.Step != StepEndCombat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	if got := counterOn(g, pw, CounterLoyalty); got != 2 {
		t.Errorf("loyalty after a 3-power attacker connected = %d, want 2", got)
	}
	if defenderSeat.Life != lifeBefore {
		t.Errorf("the planeswalker's controller lost %d life; an attack on a walker does not touch the player",
			lifeBefore-defenderSeat.Life)
	}
}

// TestPlaneswalkerControllerOwesABlockDecision is the half that is
// easiest to miss and worst to get wrong: #328's auto-pass guard
// reads SeatOwesBlockDecision, so a seat that appeared to owe nothing
// would have its one chance to block passed for it.
func TestPlaneswalkerControllerOwesABlockDecision(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := g.Seats[g.Turn.ActiveSeat].ID
	defenderSeat := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pw := pushPlaneswalkerForTest(g, defenderSeat.ID, "Their Walker", 5)
	bear := pushReadyAttackerForTest(g, attacker, "Bear", 3)
	pushReadyAttackerForTest(g, defenderSeat.ID, "Their Blocker", 2)
	stepToDeclareAttackers(t, g)
	if err := g.DeclareAttacker(bear, pw); err != nil {
		t.Fatalf("declare: %v", err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.Step != StepDeclareBlockers {
		t.Fatalf("step = %s, want declare_blockers", g.Turn.Step)
	}

	if !g.SeatOwesBlockDecision(defenderSeat.ID) {
		t.Error("the planeswalker's controller was not offered a block decision")
	}
}

// TestAttackTargetsForEffectListsPlayersWalkersAndBattles is the set
// the client's picker and the bot's enumerator both read.
func TestAttackTargetsForEffectListsPlayersWalkersAndBattles(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	them := g.Seats[1].ID
	third := g.Seats[2].ID
	mine := pushPlaneswalkerForTest(g, me, "My Walker", 4)
	theirs := pushPlaneswalkerForTest(g, them, "Their Walker", 4)
	protectedByMe := pushBattleForTest(g, them, me, "I defend this", 5)
	attackable := pushBattleForTest(g, them, third, "Someone else defends this", 5)

	var players, walkers, battles int
	seen := map[uuid.UUID]bool{}
	g.WithWriteLock(func() {
		for _, t := range g.AttackTargetsForEffect(me) {
			seen[t.ID] = true
			switch t.Kind {
			case AttackTargetPlayer:
				players++
			case AttackTargetPlaneswalker:
				walkers++
			case AttackTargetBattle:
				battles++
			}
		}
	})

	if players != len(g.Seats)-1 {
		t.Errorf("player targets = %d, want %d", players, len(g.Seats)-1)
	}
	if walkers != 1 || !seen[theirs] {
		t.Errorf("planeswalker targets = %d; want exactly the opponent's", walkers)
	}
	if seen[mine] {
		t.Error("your own planeswalker was offered as an attack target")
	}
	if battles != 1 || !seen[attackable] {
		t.Errorf("battle targets = %d; want exactly the one you don't protect", battles)
	}
	if seen[protectedByMe] {
		t.Error("a battle you protect was offered as an attack target")
	}
}
