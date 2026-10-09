package protocol

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// designation_view_test.go — ADR 0071 decision 4. Two fields, both
// public, both present only when they mean something.
//
// The redaction half (a non-knower gets neither, because a level says
// "Class" and a solved flag says "Case" as loudly as loyalty says
// "planeswalker") is covered by the allowlist table in
// face_down_view_test.go, which fails if a new CardView field escapes
// it.

func TestDesignationViewStampsLevelAndSolved(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	classID, levelledID, caseID, plainID, inHandID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UnixNano()
	// Battlefield cards are public, but a card pushed straight onto a
	// zone has an empty knower set, and FilterViewFor redacts for a
	// non-knower — including these two fields (a level says "Class").
	seen := map[uuid.UUID]bool{owner.ID: true, g.Seats[1].ID: true}

	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: classID, Name: "Wizard Class", TypeLine: "Enchantment — Class",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: levelledID, Name: "Levelled Class", TypeLine: "Enchantment — Class",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, ClassLevel: 3, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: caseID, Name: "Solved Case", TypeLine: "Enchantment — Case",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, Solved: true, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: plainID, Name: "Sol Ring", TypeLine: "Artifact",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
		// CR 400.7 cleared the designation on the way out, and the
		// derived "level 1" must not appear in its place off the
		// battlefield.
		owner.Hand.PushTop(game.Card{
			InstanceID: inHandID, Name: "Wizard Class", TypeLine: "Enchantment — Class",
			Owner: owner.ID, Controller: owner.ID, KnownBy: seen,
		})
	})

	v := ViewOfGameFor(g, owner.ID.String())
	find := func(id uuid.UUID) CardView {
		t.Helper()
		for _, c := range v.Battlefield.Cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
		for _, z := range v.Seats {
			for _, c := range z.Hand.Cards {
				if c.InstanceID == id.String() {
					return c
				}
			}
		}
		t.Fatalf("card %s missing from the view", id)
		return CardView{}
	}

	if got := find(classID).ClassLevel; got != 1 {
		t.Errorf("an unlevelled Class on the battlefield projects level %d, want 1 (CR 716.2b)", got)
	}
	if got := find(levelledID).ClassLevel; got != 3 {
		t.Errorf("a level-3 Class projects level %d, want 3", got)
	}
	if got := find(plainID).ClassLevel; got != 0 {
		t.Errorf("a permanent that is not a Class projects level %d, want absent", got)
	}
	if got := find(inHandID).ClassLevel; got != 0 {
		t.Errorf("a Class in hand projects level %d, want absent — it has no designation off the battlefield", got)
	}
	if !find(caseID).Solved {
		t.Error("a solved Case must project solved")
	}
	if find(classID).Solved || find(plainID).Solved {
		t.Error("solved must be absent for anything that is not a solved Case")
	}
}

// TestMonstrousViewIsPublicOnTheBattlefield — #1700: CardView.monstrous
// is read straight off a battlefield permanent, like harnessed, and is
// absent for one that is not monstrous.
func TestMonstrousViewIsPublicOnTheBattlefield(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	monsterID, plainID := uuid.New(), uuid.New()
	now := time.Now().UnixNano()
	seen := map[uuid.UUID]bool{owner.ID: true, g.Seats[1].ID: true}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: monsterID, Name: "Stormbreath Dragon", TypeLine: "Creature — Dragon",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, Monstrous: true, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: plainID, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
	})
	v := ViewOfGameFor(g, g.Seats[1].ID.String())
	got := map[string]bool{}
	for _, c := range v.Battlefield.Cards {
		got[c.InstanceID] = c.Monstrous
	}
	if !got[monsterID.String()] {
		t.Error("a monstrous permanent does not project monstrous to an opponent")
	}
	if got[plainID.String()] {
		t.Error("a permanent that is not monstrous projects monstrous")
	}
}

// TestRenownedViewIsPublicOnTheBattlefield — #2049: CardView.renowned
// is read straight off a battlefield permanent, like monstrous, and is
// absent for one that is not renowned.
func TestRenownedViewIsPublicOnTheBattlefield(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	renownedID, plainID := uuid.New(), uuid.New()
	now := time.Now().UnixNano()
	seen := map[uuid.UUID]bool{owner.ID: true, g.Seats[1].ID: true}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: renownedID, Name: "Topan Freeblade", TypeLine: "Creature — Human Soldier",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, Renowned: true, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: plainID, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
	})
	v := ViewOfGameFor(g, g.Seats[1].ID.String())
	got := map[string]bool{}
	for _, c := range v.Battlefield.Cards {
		got[c.InstanceID] = c.Renowned
	}
	if !got[renownedID.String()] {
		t.Error("a renowned permanent does not project renowned to an opponent")
	}
	if got[plainID.String()] {
		t.Error("a permanent that is not renowned projects renowned")
	}
}

// TestSuspectedViewIsPublicOnTheBattlefield — #2698: CardView.suspected
// is read straight off a battlefield permanent, like monstrous, is
// absent for one that is not suspected, and survives a face-down
// permanent: the designation was given to the object in public, and
// the table needs it to know why the creature cannot block.
func TestSuspectedViewIsPublicOnTheBattlefield(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	suspectID, plainID, downID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UnixNano()
	seen := map[uuid.UUID]bool{owner.ID: true, g.Seats[1].ID: true}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: suspectID, Name: "Barbed Servitor", TypeLine: "Artifact Creature — Construct",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, Suspected: true, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: plainID, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: downID, Name: "Hidden", TypeLine: "Creature — Bear", FaceDown: true,
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, Suspected: true,
			KnownBy: map[uuid.UUID]bool{owner.ID: true},
		})
	})
	v := ViewOfGameFor(g, g.Seats[1].ID.String())
	got := map[string]bool{}
	for _, c := range v.Battlefield.Cards {
		got[c.InstanceID] = c.Suspected
	}
	if !got[suspectID.String()] {
		t.Error("a suspected permanent does not project suspected to an opponent")
	}
	if got[plainID.String()] {
		t.Error("a permanent that is not suspected projects suspected")
	}
	if !got[downID.String()] {
		t.Error("a suspected face-down permanent hides its designation from an opponent")
	}
}
