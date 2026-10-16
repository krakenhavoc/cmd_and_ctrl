package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ixidron_entry_test.go — #3862: Ixidron's sweep is part of its entry
// (CR 614.1c), not something that happens just after it.

const ixidronOracle = "a86587ea-14ec-45db-8bac-10b6b9eaeb25"

// Ixidron's ruling: "There is never a moment when Ixidron is on the
// battlefield and the creatures are face-up. If a creature on the
// battlefield has a 'whenever another creature enters' ability, it
// won't trigger because that creature will be face down before Ixidron
// enters." A Soul Warden already on the battlefield is turned over as
// part of the entry, so it never sees Ixidron arrive.
func TestIxidronTurnsCreaturesDownBeforeTheyCanSeeItEnter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	warden := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soul Warden", OracleID: "f3fad295-1af2-4ecc-8546-b121ad6be27b",
		TypeLine: "Creature — Human Cleric", ManaCost: "{W}", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	life := me.Life

	castCatalogSpell(t, g, "Ixidron", "Creature — Illusion", ixidronOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	if !faceDownCardForTest(t, g, warden).FaceDown {
		t.Fatal("Soul Warden is still face up")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d: the Warden was face down before Ixidron entered and has nothing to trigger", me.Life, life)
	}
}

// Ixidron's other ruling: "If Ixidron and another creature are entering
// at the same time, the other creature enters face up." The creature
// that was already there is still turned over.
func TestIxidronLeavesACreatureEnteringWithItFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	already := pushCreatureToBattlefieldForTest(g, me.ID, "Already Here")
	ixidron := handCardForTest(me, "Ixidron", "Creature — Illusion", ixidronOracle)
	bearID := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bearID, Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var entered []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.PutOntoBattlefieldTogetherThenForEffect(
			[]game.BatchEntry{{CardID: ixidron, From: game.ZoneHand}, {CardID: bearID, From: game.ZoneHand}},
			game.ZoneEntryOptions{}, func(_ *game.Game, ids []uuid.UUID) error {
				entered = ids
				return nil
			}); err != nil {
			t.Fatalf("enter together: %v", err)
		}
	})
	if len(entered) != 2 {
		t.Fatalf("entered %v, want both", entered)
	}
	if !faceDownCardForTest(t, g, already).FaceDown {
		t.Error("the creature that was already on the battlefield is face up")
	}
	if faceDownCardForTest(t, g, entered[1]).FaceDown {
		t.Error("the creature entering with Ixidron was turned face down; it enters face up")
	}
}
