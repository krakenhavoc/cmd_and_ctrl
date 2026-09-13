package game

import (
	"testing"

	"github.com/google/uuid"
)

// battle_test.go — S27, CR 310. The lifecycle keys off the card TYPE,
// not off the catalog, so every test here uses a bare battle Card with
// a printed StartingDefense and no Spec at all. That is the property
// worth pinning: a battle a player imported from a decklist works
// without anyone having written a card file for it.

// enterBattleForTest moves a battle from a player's hand onto the
// battlefield through the real entry path, so the ETB hook fires.
func enterBattleForTest(t *testing.T, g *Game, owner uuid.UUID, name string, defense int) uuid.UUID {
	t.Helper()
	var p *Player
	for _, seat := range g.Seats {
		if seat.ID == owner {
			p = seat
		}
	}
	if p == nil {
		t.Fatalf("no such seat %v", owner)
	}
	id := uuid.New()
	p.Hand.PushTop(Card{
		InstanceID:      id,
		Name:            name,
		TypeLine:        "Battle — Siege",
		Owner:           owner,
		Controller:      owner,
		StartingDefense: defense,
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: owner}, ZoneRef{Kind: ZoneBattlefield}, id); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	return id
}

func protectorChoiceFor(g *Game, chooser uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceChooseProtector && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func protectorOf(g *Game, battleID uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == battleID {
			return c.ProtectorPlayerID
		}
	}
	return uuid.Nil
}

// TestBattleEntersWithItsPrintedDefense is the #274 lesson applied to
// battles: printed data on the card, stamped by the engine, so a
// battle nobody wrote a Spec for is playable. Before this, every
// battle entered with zero defense counters and the CR 704.5p
// state-based action swept it into the graveyard on the next priority
// boundary.
func TestBattleEntersWithItsPrintedDefense(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	b := enterBattleForTest(t, g, owner, "Test Siege", 5)

	if got := counterOn(g, b, CounterDefense); got != 5 {
		t.Fatalf("defense counters on entry = %d, want 5", got)
	}
}

// TestBattleQueuesAProtectorChoiceForItsController is CR 310.5. The
// controller chooses; the options are their opponents and nobody
// else.
func TestBattleQueuesAProtectorChoiceForItsController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	b := enterBattleForTest(t, g, owner, "Test Siege", 5)

	ch := protectorChoiceFor(g, owner)
	if ch == nil {
		t.Fatal("no protector choice queued for the battle's controller")
	}
	if ch.Source != b {
		t.Errorf("choice source = %v, want the battle", ch.Source)
	}
	if len(ch.PickTargetPlayers) != len(g.Seats)-1 {
		t.Errorf("protector options = %d, want %d (every opponent)", len(ch.PickTargetPlayers), len(g.Seats)-1)
	}
	for _, id := range ch.PickTargetPlayers {
		if id == owner {
			t.Error("the battle's controller was offered as its own protector")
		}
	}
}

// TestResolveChooseProtectorRecordsTheProtector, and the protector is
// what decides who defends it (CR 310.7) — which
// defendingPlayerForAttackLocked reads and the attack gate enforces.
func TestResolveChooseProtectorRecordsTheProtector(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	want := g.Seats[2].ID
	b := enterBattleForTest(t, g, owner, "Test Siege", 5)

	ch := protectorChoiceFor(g, owner)
	if ch == nil {
		t.Fatal("no protector choice queued")
	}
	if err := g.ResolveChooseProtector(ch.ID, owner, want); err != nil {
		t.Fatalf("ResolveChooseProtector: %v", err)
	}
	if got := protectorOf(g, b); got != want {
		t.Errorf("protector = %v, want %v", got, want)
	}
	if protectorChoiceFor(g, owner) != nil {
		t.Error("the protector choice was not drained")
	}

	var defender uuid.UUID
	g.WithWriteLock(func() { defender = g.defendingPlayerForAttackLocked(b) })
	if defender != want {
		t.Errorf("defending player for the battle = %v, want the PROTECTOR %v", defender, want)
	}
}

// TestResolveChooseProtectorRejectsANonOpponent — the controller may
// not protect their own battle, and a seat that was never offered is
// not a legal answer.
func TestResolveChooseProtectorRejectsANonOpponent(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	enterBattleForTest(t, g, owner, "Test Siege", 5)

	ch := protectorChoiceFor(g, owner)
	if ch == nil {
		t.Fatal("no protector choice queued")
	}
	if err := g.ResolveChooseProtector(ch.ID, owner, owner); err == nil {
		t.Error("the controller was accepted as their own battle's protector")
	}
	if err := g.ResolveChooseProtector(ch.ID, g.Seats[1].ID, g.Seats[2].ID); err == nil {
		t.Error("a player who is not the chooser answered the protector prompt")
	}
}

// TestBattleAtZeroDefenseIsSweptAndAnnounced covers both halves of a
// defeat: the CR 310.9 announcement that a defeated trigger harvests,
// and the CR 704.5p sweep that follows it in the same pass.
func TestBattleAtZeroDefenseIsSweptAndAnnounced(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	b := enterBattleForTest(t, g, owner, "Test Siege", 2)
	if ch := protectorChoiceFor(g, owner); ch != nil {
		if err := g.ResolveChooseProtector(ch.ID, owner, g.Seats[1].ID); err != nil {
			t.Fatalf("ResolveChooseProtector: %v", err)
		}
	}

	cut := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, b, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
		g.runStateChecksLocked()
	})

	defeated := false
	for _, ev := range g.Events[cut:] {
		if ev.Kind == EventBattleDefeated && ev.CardID == b {
			defeated = true
		}
	}
	if !defeated {
		t.Error("no battle_defeated event was emitted")
	}
	if counterOn(g, b, CounterDefense) >= 0 {
		t.Error("a battle at zero defense survived the CR 704.5p sweep")
	}
}

// TestProtectorIsClearedWhenTheBattleLeaves is CR 400.7: a battle that
// returns is a new object and picks a new protector. Keeping the old
// one would leave a returning battle defended by whoever happened to
// be chosen last time — including, after an elimination, nobody.
func TestProtectorIsClearedWhenTheBattleLeaves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	b := enterBattleForTest(t, g, owner, "Test Siege", 5)
	ch := protectorChoiceFor(g, owner)
	if ch == nil {
		t.Fatal("no protector choice queued")
	}
	if err := g.ResolveChooseProtector(ch.ID, owner, g.Seats[1].ID); err != nil {
		t.Fatalf("ResolveChooseProtector: %v", err)
	}

	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: owner}, b); err != nil {
		t.Fatalf("move battle off the battlefield: %v", err)
	}
	for _, c := range g.Seats[0].Graveyard.Cards {
		if c.InstanceID == b && c.ProtectorPlayerID != uuid.Nil {
			t.Errorf("protector survived the battle leaving the battlefield: %v", c.ProtectorPlayerID)
		}
	}
}
