package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_view_test.go — the wire half of ADR 0103: a Room's doors, the
// unlock rows naming their door, the door log lines (owner decision 7:
// "fully unlocked" is its own line), a fully locked Room's empty name
// with its halves still on `faces` (owner decision 6), and the fused
// announce block on a fuse card in hand.

func viewSplitCard(owner uuid.UUID, oracle string, left, right game.Face) game.Card {
	c := game.NewCard(left.Name, owner)
	c.OracleID = oracle
	c.Layout = game.LayoutSplit
	c.Faces = []game.Face{left, right}
	c.SettleImported()
	return c
}

func TestRoomViewShipsDoorsUnlockRowsAndLogLines(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	room := viewSplitCard(me.ID, "view-room-oracle",
		game.Face{Name: "Dim Hall", TypeLine: "Enchantment — Room", ManaCost: "{1}{B}"},
		game.Face{Name: "Deep Cellar", TypeLine: "Enchantment — Room", ManaCost: "{4}{B}"},
	)
	g.WithWriteLock(func() {
		me.Hand.PushTop(room)
		if _, err := game.MoveCard(me.Hand, g.Battlefield, room.InstanceID); err != nil {
			t.Fatal(err)
		}
		// The entry stamp every real battlefield entry writes, which
		// is the view's battlefield test for designations.
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == room.InstanceID {
				g.Battlefield.Cards[i].EnteredBattlefieldAt = 1
				g.Battlefield.Cards[i].KnownBy = map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
			}
		}
	})

	v := cardViewByID(t, g, room.InstanceID)
	if v.Doors == nil || v.Doors.Left || v.Doors.Right {
		t.Fatalf("doors = %+v, want both present and locked", v.Doors)
	}
	if v.Name != "" {
		t.Errorf("a fully locked Room's name = %q, want empty (CR 709.5)", v.Name)
	}
	if len(v.Faces) != 2 || v.Faces[0].Name != "Dim Hall" || v.Faces[1].Name != "Deep Cellar" {
		t.Errorf("faces = %+v, want both halves for the client to label it from", v.Faces)
	}
	doors := map[string]string{}
	for _, sa := range v.SpecialActions {
		if sa.Kind == "unlock" {
			doors[sa.Door] = sa.Cost
		}
	}
	if doors["left"] != "{1}{B}" || doors["right"] != "{4}{B}" {
		t.Errorf("unlock rows = %v, want one per door priced at its half", doors)
	}

	g.WithWriteLock(func() {
		if err := g.UnlockDoorForEffect(room.InstanceID, game.DoorLeft, me.ID); err != nil {
			t.Fatal(err)
		}
		if err := g.UnlockDoorForEffect(room.InstanceID, game.DoorRight, me.ID); err != nil {
			t.Fatal(err)
		}
	})
	v = cardViewByID(t, g, room.InstanceID)
	if v.Doors == nil || !v.Doors.Left || !v.Doors.Right || v.Name != "Dim Hall // Deep Cellar" {
		t.Fatalf("after unlocking both: doors %+v, name %q", v.Doors, v.Name)
	}
	gv := ViewOfGameFor(g, me.ID.String())
	unlocked := logOfKind(gv, LogDoorUnlocked)
	if len(unlocked) != 2 || !strings.Contains(unlocked[0].Text, "unlocked Dim Hall") {
		t.Errorf("door_unlocked lines = %+v", unlocked)
	}
	full := logOfKind(gv, LogRoomFullyUnlocked)
	if len(full) != 1 || !strings.Contains(full[0].Text, "is fully unlocked") {
		t.Errorf("room_fully_unlocked lines = %+v, want one (owner decision 7)", full)
	}

	g.WithWriteLock(func() {
		if err := g.LockDoorForEffect(room.InstanceID, game.DoorLeft, me.ID); err != nil {
			t.Fatal(err)
		}
	})
	if locked := logOfKind(ViewOfGameFor(g, me.ID.String()), LogDoorLocked); len(locked) != 1 {
		t.Errorf("door_locked lines = %+v, want one", locked)
	}
}

func TestFuseCardInHandShipsAFusedBlock(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	c := viewSplitCard(me.ID, "view-fuse-oracle",
		game.Face{Name: "Snap", TypeLine: "Instant", ManaCost: "{R}", OracleText: "Snap.\nFuse (You may cast one or both halves of this card from your hand.)"},
		game.Face{Name: "Crackle", TypeLine: "Instant", ManaCost: "{1}{U}", OracleText: "Crackle.\nFuse (You may cast one or both halves of this card from your hand.)"},
	)
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	v := ViewOfGameFor(g, me.ID.String())
	var got *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == c.InstanceID.String() {
			got = &v.Seats[0].Hand.Cards[i]
		}
	}
	if got == nil {
		t.Fatal("the fuse card is not in its owner's hand view")
	}
	if got.Name != "Snap // Crackle" {
		t.Errorf("name in hand = %q, want the whole card (CR 709.4)", got.Name)
	}
	if got.Fused == nil || got.Fused.Name != "Snap // Crackle" || got.Fused.ManaCost != "{R}{1}{U}" {
		t.Fatalf("fused block = %+v, want both halves' name and cost", got.Fused)
	}
	plain := viewSplitCard(me.ID, "view-plain-split-oracle",
		game.Face{Name: "Up", TypeLine: "Instant", ManaCost: "{W}"},
		game.Face{Name: "Down", TypeLine: "Instant", ManaCost: "{B}"},
	)
	plain.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.WithWriteLock(func() { me.Hand.PushTop(plain) })
	v = ViewOfGameFor(g, me.ID.String())
	for _, hc := range v.Seats[0].Hand.Cards {
		if hc.InstanceID == plain.InstanceID.String() && hc.Fused != nil {
			t.Error("a split card without fuse ships a fused block")
		}
	}
}
