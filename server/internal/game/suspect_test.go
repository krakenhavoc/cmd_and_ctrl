package game

import (
	"testing"

	"github.com/google/uuid"
)

// suspect_test.go — CR 701.60, the suspected designation (ADR 0071
// amendment 2026-10-08, #2698): the grant, the lifetime, and the two
// ways the designation must NOT travel (a copy, a restore that
// forgets it).

func suspectedCard(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	return layeredBattlefieldCard(t, g, id)
}

func suspectNow(t *testing.T, g *Game, id uuid.UUID) bool {
	t.Helper()
	var did bool
	g.WithWriteLock(func() { did = g.SuspectForEffect(id) })
	return did
}

// A suspected creature has menace and can't block (CR 701.60c), and
// stops having either the moment it is no longer suspected.
func TestSuspectGrantsMenaceAndCantBlockUntilUnsuspected(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Suspect", TypeLine: "Creature — Human", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})

	c := suspectedCard(t, g, id)
	if HasKeyword(&c, "menace") || Restricted(&c, CantBlock) {
		t.Fatal("fixture is wrong: a fresh creature has menace or can't block")
	}

	if !suspectNow(t, g, id) {
		t.Fatal("SuspectForEffect = false for an unsuspected creature")
	}
	c = suspectedCard(t, g, id)
	if !c.Suspected {
		t.Fatal("Card.Suspected = false after suspecting")
	}
	if !HasKeyword(&c, "menace") {
		t.Errorf("a suspected creature has no menace: abilities %v", c.Effective().Abilities)
	}
	if !Restricted(&c, CantBlock) {
		t.Error("a suspected creature can block")
	}

	var undid bool
	g.WithWriteLock(func() { undid = g.UnsuspectForEffect(id) })
	if !undid {
		t.Fatal("UnsuspectForEffect = false for a suspected creature")
	}
	c = suspectedCard(t, g, id)
	if c.Suspected || HasKeyword(&c, "menace") || Restricted(&c, CantBlock) {
		t.Errorf("still suspected after unsuspecting: suspected %v, menace %v, can't block %v",
			c.Suspected, HasKeyword(&c, "menace"), Restricted(&c, CantBlock))
	}
}

// CR 701.60d: a permanent that is already suspected can't become
// suspected again — which is also what keeps its timestamp from being
// rewritten by a second suspect.
func TestSuspectingASuspectedCreatureDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Suspect", TypeLine: "Creature — Human", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})
	if !suspectNow(t, g, id) {
		t.Fatal("first suspect refused")
	}
	first := suspectedCard(t, g, id).SuspectedAt
	if first == 0 {
		t.Fatal("SuspectedAt was not stamped")
	}
	if suspectNow(t, g, id) {
		t.Error("a suspected creature became suspected again (CR 701.60d)")
	}
	if got := suspectedCard(t, g, id).SuspectedAt; got != first {
		t.Errorf("SuspectedAt changed from %d to %d on a refused second suspect", first, got)
	}
}

// Only a creature on the battlefield can be suspected; anything else
// is a quiet no-op, because the target leaving in response is
// ordinary play.
func TestSuspectRefusesWhatIsNotACreatureOnTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	rock := pushTypedTestCard(g, Card{
		Name: "Rock", TypeLine: "Artifact", Owner: seat, Controller: seat,
	})
	if suspectNow(t, g, rock) {
		t.Error("a non-creature became suspected")
	}
	if suspectNow(t, g, uuid.New()) {
		t.Error("a permanent that is not there became suspected")
	}
	var was bool
	g.ReadSnapshot(func() { was = g.IsSuspected(rock) })
	if was {
		t.Error("IsSuspected(artifact) = true")
	}
}

// CR 400.7: a suspected creature that leaves the battlefield is a new
// object when it comes back, and is not suspected.
func TestSuspectedClearsWhenThePermanentLeaves(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Suspect", TypeLine: "Creature — Human", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})
	suspectNow(t, g, id)
	var moved Card
	g.WithWriteLock(func() {
		var err error
		moved, err = MoveCard(g.Battlefield, g.Seats[0].Graveyard, id)
		if err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if moved.Suspected || moved.SuspectedAt != 0 {
		t.Errorf("the designation survived the zone change: suspected %v, at %d", moved.Suspected, moved.SuspectedAt)
	}
}

// A suspected creature is not a copiable value (CR 707.2): a Clone of
// one is a plain creature.
func TestSuspectedIsNotACopiableValue(t *testing.T) {
	source := Card{
		InstanceID: uuid.New(), Name: "Suspect", TypeLine: "Creature — Human",
		Power: 2, Toughness: 2, Suspected: true, SuspectedAt: 42,
	}
	copied := CopiableValuesOf(source)
	clone := Card{InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter"}
	clone.applyCopy(copied, source)
	if clone.Suspected || clone.SuspectedAt != 0 {
		t.Errorf("a copy took the suspected designation: %v at %d", clone.Suspected, clone.SuspectedAt)
	}
}

// The undo and the deploy: "not suspected" is a legal zero value, so a
// restore that dropped the designation would hand a suspected creature
// its blocking back and say nothing.
func TestSuspectedSurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Suspect", TypeLine: "Creature — Human", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})
	suspectNow(t, g, id)
	at := suspectedCard(t, g, id).SuspectedAt

	find := func(cards []Card) Card {
		for _, c := range cards {
			if c.InstanceID == id {
				return c
			}
		}
		return Card{}
	}
	if c := find(g.Clone().Battlefield.Cards); !c.Suspected || c.SuspectedAt != at {
		t.Errorf("clone lost the designation: %v at %d (want %d)", c.Suspected, c.SuspectedAt, at)
	}
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	var back Card
	restored.ReadSnapshot(func() { back = find(restored.Battlefield.Cards) })
	if !back.Suspected || back.SuspectedAt != at {
		t.Errorf("snapshot lost the designation: %v at %d (want %d)", back.Suspected, back.SuspectedAt, at)
	}
	if c := layeredBattlefieldCard(t, restored, id); !Restricted(&c, CantBlock) {
		t.Error("the restored suspected creature can block")
	}
}

// The block gate reads the grant: a suspected creature is refused as a
// blocker, and an unsuspected one beside it is not.
func TestSuspectedCreatureCannotBeDeclaredAsABlocker(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
	suspect := pushCombatant(t, g, g.Seats[1], "Suspect", 2, 2)
	free := pushCombatant(t, g, g.Seats[1], "Free", 2, 2)
	suspectNow(t, g, suspect)
	declareAttacks(t, g, attacker)

	if err := g.DeclareBlocker(suspect, attacker); err == nil {
		t.Fatal("a suspected creature was allowed to block (CR 701.60c)")
	}
	if err := g.DeclareBlocker(free, attacker); err != nil {
		t.Fatalf("an unsuspected creature was refused: %v", err)
	}
}

// Layer 6 ordering (CR 613.7): the menace grant is ordered at the
// moment the creature became suspected. A "loses all abilities" that
// is OLDER leaves the menace; one that is NEWER takes it away. The
// can't-block half is a restriction and no layer touches it, which is
// the declared, stricter-than-printed reading in suspect.go.
func TestSuspectMenaceIsOrderedByTimestampAgainstAbilityRemoval(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var victim uuid.UUID
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == silencerOracle {
			return silencerStatics(&victim)
		}
		return nil
	})

	victim = pushTypedTestCard(g, Card{
		Name: "Victim", TypeLine: "Creature — Human", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})

	// Older removal: it entered first, so the suspect grant lands after.
	pushTypedTestCard(g, Card{
		Name: "Old Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
		Owner: seat, Controller: seat,
	})
	suspectNow(t, g, victim)
	c := suspectedCard(t, g, victim)
	if !HasKeyword(&c, "menace") {
		t.Error("a removal older than the designation took the menace away")
	}
	if !Restricted(&c, CantBlock) {
		t.Error("can't block is missing")
	}
}

func TestSuspectMenaceIsRemovedByANewerAbilityRemoval(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var victim uuid.UUID
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == silencerOracle {
			return silencerStatics(&victim)
		}
		return nil
	})
	victim = pushTypedTestCard(g, Card{
		Name: "Victim", TypeLine: "Creature — Human", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})
	suspectNow(t, g, victim)
	// Entered after the designation: its removal sorts later.
	pushTypedTestCard(g, Card{
		Name: "New Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
		Owner: seat, Controller: seat,
	})
	c := suspectedCard(t, g, victim)
	if HasKeyword(&c, "menace") {
		t.Error("a removal newer than the designation left the menace on")
	}
	if !c.Suspected {
		t.Error("losing abilities un-suspected the creature; it should stay suspected")
	}
}
