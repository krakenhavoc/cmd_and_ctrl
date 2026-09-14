package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// battles_test.go — S27, CR 310, against the real catalog.
//
// The lifecycle itself is pinned in server/internal/game/battle_test.go
// against a bare battle Card with no Spec, which is the property that
// matters most. What is tested here is the catalog half: the fallback
// defense, and a real Siege's ETB and defeated triggers.

const invasionOfInnistradOracle = invasionOfInnistradOracleID

// TestBattleSpecFallbackStampsDefense covers the path a card that
// never went through deck import takes — a fixture, a token, the demo
// seed. The Card carries no StartingDefense, so the engine falls back
// to the catalog's BattleSpec rather than letting the CR 704.5p
// state-based action eat the battle on arrival.
func TestBattleSpecFallbackStampsDefense(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	owner.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Invasion of Innistrad",
		OracleID:   invasionOfInnistradOracle,
		TypeLine:   "Battle — Siege",
		Owner:      owner.ID,
		Controller: owner.ID,
		// StartingDefense deliberately ZERO: this is the
		// never-imported case, and the catalog is the only source.
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}

	got := -1
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			got = c.Counters[game.CounterDefense]
		}
	}
	if got != 5 {
		t.Fatalf("defense counters from the catalog fallback = %d, want 5", got)
	}
}

// TestInvasionOfInnistradShrinksACreatureAndIsDefeated walks the card
// end to end: the ETB removal, and the defeated trigger exiling the
// Siege once its defense is gone.
func TestInvasionOfInnistradShrinksACreatureAndIsDefeated(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	victimSeat := g.Seats[(seat+1)%len(g.Seats)]
	victim := pushCreatureToBattlefieldForTest(g, victimSeat.ID, "Doomed Bear")

	id := uuid.New()
	owner.Hand.PushTop(game.Card{
		InstanceID:      id,
		Name:            "Invasion of Innistrad",
		OracleID:        invasionOfInnistradOracle,
		TypeLine:        "Battle — Siege",
		Owner:           owner.ID,
		Controller:      owner.ID,
		StartingDefense: 5,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}

	// The protector prompt blocks priority; answer it first.
	answerProtectorPrompt(t, g, owner.ID, victimSeat.ID)
	// Then the ETB trigger's target pick, then let it resolve.
	answerFirstPickTarget(t, g)
	passPriorityAroundTable(t, g)

	if onBattlefield(g, victim) {
		t.Error("the -13/-13 creature survived; the CR 704.5f SBA should have taken it")
	}

	// Now defeat the battle: five damage takes the last defense
	// counter (CR 120.3e), and the defeated trigger exiles it.
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, 5); err != nil {
			t.Fatalf("damage the battle: %v", err)
		}
	})
	// The state-based actions fire at the next priority BOUNDARY, and
	// with four seats a single PassPriority only rotates — it does not
	// wrap, so it runs no check. AdvanceStep crosses a boundary
	// unconditionally, which is the cheapest way to get there, and the
	// defeated trigger drains onto the stack from the same pass.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	passPriorityAroundTable(t, g)

	if onBattlefield(g, id) {
		t.Fatal("a battle at zero defense survived")
	}
	if !inExile(g, id) {
		t.Error("the defeated Siege was not exiled")
	}
}

func inExile(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func answerProtectorPrompt(t *testing.T, g *game.Game, chooser, protector uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseProtector && c.Chooser == chooser {
			if err := g.ResolveChooseProtector(c.ID, chooser, protector); err != nil {
				t.Fatalf("ResolveChooseProtector: %v", err)
			}
			return
		}
	}
	t.Fatal("no protector prompt was queued")
}

func answerFirstPickTarget(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoicePickTarget {
			continue
		}
		if len(c.PickTargetCards) == 0 {
			t.Fatalf("pick_target prompt %q offered no cards", c.Reason)
		}
		pick := game.TargetRef{Kind: game.TargetCard, ID: c.PickTargetCards[0]}
		if err := g.ResolvePickTarget(c.ID, c.Chooser, pick); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		return
	}
	t.Fatal("no pick_target prompt was queued")
}
