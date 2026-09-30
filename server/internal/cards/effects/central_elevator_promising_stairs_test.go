package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	centralElevatorOracle = "927fa223-69a3-4563-80b1-b578f3d031f1"
	roomsDLibraryOracle   = "rooms-d-library-fixture"
)

func centralElevatorCard(owner uuid.UUID) game.Card {
	return roomsDCard(owner, centralElevatorOracle, "Central Elevator", "{3}{U}", "Promising Stairs", "{2}{U}", "U")
}

// Casting the left half unlocks it on entry (CR 709.5d, 709.5h) and the
// search skips a library Room that shares a name with a Room you
// control (CR 201.2): the decoy has "Central Elevator" as one of its
// two names.
func TestCentralElevatorFetchesARoomWithANewName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	fresh := roomsDCard(me.ID, roomsDLibraryOracle, "Fresh Hall", "{1}{W}", "Fresh Vault", "{3}{W}", "W")
	decoy := roomsDCard(me.ID, roomsDLibraryOracle, "Elsewhere", "{1}{W}", "Central Elevator", "{3}{W}", "W")
	me.Library.PushTop(fresh)
	me.Library.PushTop(decoy)

	roomsDCast(t, g, me, centralElevatorCard(me.ID), 0)

	if !me.Hand.Contains(fresh.InstanceID) {
		t.Fatal("the Room with a new name did not reach the hand")
	}
	if me.Hand.Contains(decoy.InstanceID) || !me.Library.Contains(decoy.InstanceID) {
		t.Error("a Room sharing the name Central Elevator was fetched")
	}
}

func promisingStairsUpkeepFor(t *testing.T, g *game.Game, me *game.Player, stairs uuid.UUID) {
	t.Helper()
	item := &game.StackItem{Controller: me.ID, SourceCardID: stairs}
	g.WithWriteLock(func() {
		if err := promisingStairsUpkeep(g, item); err != nil {
			t.Fatal(err)
		}
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil && c.Chooser == me.ID {
			if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
				t.Fatalf("ResolveSurveil: %v", err)
			}
			return
		}
	}
}

// Promising Stairs surveils 1, then wins with eight different names among
// unlocked doors of Rooms you control; seven do not.
func TestPromisingStairsWinsAtEightDifferentNames(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rooms int
		win   bool
	}{{"seven names", 3, false}, {"eight names", 4, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			roomsDCast(t, g, me, centralElevatorCard(me.ID), 1) // Promising Stairs: one unlocked door
			stairs := findBattlefieldByName(g, "Promising Stairs")
			if stairs == uuid.Nil {
				t.Fatal("Promising Stairs is not on the battlefield")
			}
			names := []string{"A", "B", "C", "D"}
			for i := 0; i < tc.rooms; i++ {
				roomsDOnBattlefield(g, me.ID, roomsDLibraryOracle, names[i]+"1", names[i]+"2", game.DoorLeftUnlocked|game.DoorRightUnlocked)
			}
			if n := distinctDoorNames(g, me.ID); n != 2*tc.rooms+1 {
				t.Fatalf("distinct names = %d, want %d", n, 2*tc.rooms+1)
			}
			promisingStairsUpkeepFor(t, g, me, stairs)
			if won := g.State != game.StateActive; won != tc.win {
				t.Errorf("game over = %v, want %v", won, tc.win)
			}
		})
	}
}
