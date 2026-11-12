package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const phasingOfZhalfirOracle = "6b92a5f4-a241-48c8-9aec-c8e9d5b01fa5"

func zhalfirPhasedOut(g *game.Game, id uuid.UUID) bool {
	out := false
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(id) })
	return out
}

// The whole Saga, chapter by chapter: I and II each phase a permanent
// out and keep it out through its controller's untap steps; III destroys
// every creature and pays each controller in Phyrexians; and once the
// Saga is sacrificed the phased-out permanents come back at their
// controller's next untap step (CR 502.1), not before.
func TestThePhasingOfZhalfirPlaysItsThreeChapters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % 4
	opp := g.Seats[oppSeat]
	first := pushVanillaCreature(g, opp.ID, "First Prisoner", 3, 3)
	second := pushPermanentForTest(g, opp.ID, "Second Prisoner", "", "Artifact")
	land := pushPermanentForTest(g, opp.ID, "Test Land", "", "Land")

	id := castReadAheadSaga(t, g, "The Phasing of Zhalfir", phasingOfZhalfirOracle)
	chooseChapter(t, g, 1)

	// "Another target nonland permanent": not the Saga, not a land.
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("chapter I did not ask for a target")
	}
	for _, bad := range []uuid.UUID{id, land} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("chapter I accepted an illegal target (%s)", bad)
		}
	}
	pickTriggerTarget(t, g, me.ID, first)
	passPriorityAroundTable(t, g)
	if !zhalfirPhasedOut(g, first) {
		t.Fatal("chapter I did not phase its target out")
	}

	// Round the table: the opponent's untap step passes and it stays out.
	advanceToPrecombatMainOf(t, g, seat)
	if !zhalfirPhasedOut(g, first) {
		t.Fatal("it phased in while its jailer still controls the Saga")
	}
	if got := chaptersFired(g, id); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("chapters fired = %v, want [1 2]", got)
	}
	pickTriggerTarget(t, g, me.ID, second)
	passPriorityAroundTable(t, g)
	if !zhalfirPhasedOut(g, second) {
		t.Fatal("chapter II did not phase its target out")
	}

	// Creatures for chapter III: two of mine, one of the opponent's,
	// and an indestructible one that is not destroyed and pays nothing.
	pushVanillaCreature(g, me.ID, "My Bear A", 2, 2)
	pushVanillaCreature(g, me.ID, "My Bear B", 2, 2)
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	third := g.Seats[(seat+2)%4]
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Bear", TypeLine: "Artifact Creature — Bear",
		Power: 2, Toughness: 2, Owner: third.ID, Controller: third.ID,
		Keywords: []string{"indestructible"},
	})

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := chaptersFired(g, id); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("chapters fired = %v, want [1 2 3]", got)
	}
	for _, name := range []string{"My Bear A", "My Bear B", "Their Bear"} {
		if countBattlefieldByName(g, name) != 0 {
			t.Errorf("%s survived chapter III", name)
		}
	}
	if countBattlefieldByName(g, "Darksteel Bear") != 1 {
		t.Error("the indestructible creature was destroyed")
	}
	tokens := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Phyrexian" {
			continue
		}
		tokens[c.Controller]++
		if !c.IsToken() || c.Power != 2 || c.Toughness != 2 || !slices.Equal(c.Colors, []string{"B"}) ||
			c.TypeLine != "Token Creature — Phyrexian" {
			t.Errorf("token = %q %d/%d %v, want a 2/2 black Phyrexian creature token", c.TypeLine, c.Power, c.Toughness, c.Colors)
		}
	}
	if tokens[me.ID] != 2 || tokens[opp.ID] != 1 || tokens[third.ID] != 0 || len(tokens) != 2 {
		t.Errorf("Phyrexians by controller = %v, want 2 for the Saga's controller, 1 for the opponent, none for the indestructible creature's", tokens)
	}
	if g.Battlefield.Contains(id) || !inGraveyardOf(g, me.ID, id) {
		t.Fatal("the Saga was not sacrificed after chapter III")
	}
	// Phased-out permanents are untouched by the wipe (CR 702.26b), and
	// the Saga leaving phases nothing in on the spot.
	if !zhalfirPhasedOut(g, first) || !zhalfirPhasedOut(g, second) {
		t.Fatal("a permanent phased in the moment the Saga left; it waits for its controller's untap step")
	}

	advanceToPrecombatMainOf(t, g, oppSeat)
	if !g.Battlefield.Contains(first) || !g.Battlefield.Contains(second) {
		t.Fatal("the permanents did not phase in at their controller's next untap step after the Saga left")
	}
}

// The Saga destroyed with chapter I on the stack: the target still
// phases out, but the duration never starts (CR 611.2b), so it is back
// at its controller's next untap step.
func TestThePhasingOfZhalfirGoneBeforeTheChapterResolves(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % 4
	opp := g.Seats[oppSeat]
	prisoner := pushVanillaCreature(g, opp.ID, "Prisoner", 3, 3)

	id := castReadAheadSaga(t, g, "The Phasing of Zhalfir", phasingOfZhalfirOracle)
	chooseChapter(t, g, 1)
	pickTriggerTarget(t, g, me.ID, prisoner)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	passPriorityAroundTable(t, g)

	if !zhalfirPhasedOut(g, prisoner) {
		t.Fatal("the target did not phase out")
	}
	advanceToPrecombatMainOf(t, g, oppSeat)
	if !g.Battlefield.Contains(prisoner) {
		t.Fatal("with no Saga to control, it phases in at its controller's next untap step")
	}
}

// "For as long as YOU control this Saga": an opponent taking the Saga
// ends the effect, and the permanent is back at its next untap step
// while the Saga is still on the battlefield.
func TestThePhasingOfZhalfirReleasesWhenItChangesController(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % 4
	opp := g.Seats[oppSeat]
	prisoner := pushVanillaCreature(g, opp.ID, "Prisoner", 3, 3)

	id := castReadAheadSaga(t, g, "The Phasing of Zhalfir", phasingOfZhalfirOracle)
	chooseChapter(t, g, 1)
	pickTriggerTarget(t, g, me.ID, prisoner)
	passPriorityAroundTable(t, g)
	if !zhalfirPhasedOut(g, prisoner) {
		t.Fatal("the target did not phase out")
	}

	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.Nil, id, opp.ID, game.IndefiniteDuration(), "test theft") {
			t.Fatal("could not change the Saga's controller")
		}
	})
	g.RunStateChecksForTest()
	advanceToPrecombatMainOf(t, g, oppSeat)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Saga left; this test wants it still on the battlefield")
	}
	if !g.Battlefield.Contains(prisoner) {
		t.Fatal("the Saga changed controller, so the permanent phases in at its controller's next untap step")
	}
}

// A table saved with a permanent held out is a restore point: the
// restored game keeps it out, and releases it after the Saga is gone.
func TestThePhasingOfZhalfirSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % 4
	opp := g.Seats[oppSeat]
	prisoner := pushVanillaCreature(g, opp.ID, "Prisoner", 3, 3)

	id := castReadAheadSaga(t, g, "The Phasing of Zhalfir", phasingOfZhalfirOracle)
	chooseChapter(t, g, 1)
	pickTriggerTarget(t, g, me.ID, prisoner)
	passPriorityAroundTable(t, g)

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	advanceToPrecombatMainOf(t, restored, oppSeat)
	if !zhalfirPhasedOut(restored, prisoner) {
		t.Fatal("the restored game let it phase in while the Saga is still controlled")
	}
	restored.WithWriteLock(func() { restored.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	restored.RunStateChecksForTest()
	advanceToPrecombatMainOf(t, restored, oppSeat)
	if !restored.Battlefield.Contains(prisoner) {
		t.Fatal("the restored game did not release it after the Saga left")
	}
}
