package game

import (
	"testing"

	"github.com/google/uuid"
)

// game_action_replacements_test.go — #6096: the engine half of "If you
// would lose the game, instead …" and "If you would planeswalk, instead
// …", with the catalog hook stubbed. The card-level tests (The Golden
// Throne, Susan Foreman) live in cards/effects.

const (
	testThroneOracle = "test-6096-golden-throne"
	testSusanOracle  = "test-6096-susan-foreman"
)

func withCatalogGameActionReplacements(t *testing.T) {
	t.Helper()
	prev := CatalogGameActionReplacements
	CatalogGameActionReplacements = func(key string) []GameActionReplacement {
		switch key {
		case testThroneOracle:
			return []GameActionReplacement{{Kind: ReplaceLoseGameExileThisLifeBecomes, Life: 1}}
		case testSusanOracle:
			return []GameActionReplacement{{Kind: ReplacePlaneswalkOrderTopTwo}}
		}
		return nil
	}
	t.Cleanup(func() { CatalogGameActionReplacements = prev })
}

func seedReplacementPermanent(g *Game, owner *Player, oracle, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.OracleID = oracle
	c.TypeLine = "Legendary Artifact"
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// A state-based loss is replaced once: the permanent is exiled, the life
// total becomes 1, and the next loss, with the permanent gone, stands.
func TestALossReplacementAppliesOnceToAStateBasedLoss(t *testing.T) {
	withCatalogGameActionReplacements(t)
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	var throne uuid.UUID
	g.WithWriteLock(func() {
		throne = seedReplacementPermanent(g, p, testThroneOracle, "Throne")
		p.Life = -4
	})
	g.RunStateChecksForTest()
	if p.Eliminated || p.Life != 1 {
		t.Fatalf("eliminated %v at %d life, want the loss replaced at 1 life", p.Eliminated, p.Life)
	}
	if !g.Exile.Contains(throne) {
		t.Error("the replacing permanent was not exiled")
	}
	ev, ok := lastEvent(g, EventGameLossReplaced)
	if !ok || ev.Actor != p.ID || ev.CardID != throne || ev.Label != string(LossLife) {
		t.Errorf("replacement event %+v", ev)
	}
	g.WithWriteLock(func() { p.Life = 0 })
	g.RunStateChecksForTest()
	if !p.Eliminated {
		t.Error("the second loss was replaced with nothing left to replace it")
	}
}

// An effect loss is replaced; a gate is read first, so a player who
// can't lose replaces nothing; a concession is never replaced.
func TestALossReplacementCoversEffectLossesButNotConcessionsOrGatedLosses(t *testing.T) {
	withCatalogGameActionReplacements(t)
	withCatalogGameEndGates(t)
	g := newActiveGameWithSeats(t, 3)
	p, q := g.Seats[1], g.Seats[2]
	var throne uuid.UUID
	g.WithWriteLock(func() {
		throne = seedReplacementPermanent(g, p, testThroneOracle, "Throne")
		lost, err := g.LoseTheGameForEffect(p.ID, uuid.Nil)
		if lost || err != nil || p.Eliminated || p.Life != 1 {
			t.Fatalf("lost %v (%v), eliminated %v, life %d", lost, err, p.Eliminated, p.Life)
		}
	})
	if !g.Exile.Contains(throne) {
		t.Error("the replacing permanent was not exiled")
	}

	var angelThrone uuid.UUID
	g.WithWriteLock(func() {
		seedGatePermanent(g, q, testAngelOracle, "Angel")
		angelThrone = seedReplacementPermanent(g, q, testThroneOracle, "Throne")
		q.Life = 0
	})
	g.RunStateChecksForTest()
	if q.Eliminated || !g.Battlefield.Contains(angelThrone) {
		t.Error("a player who can't lose had the loss replaced anyway")
	}

	g.WithWriteLock(func() { seedReplacementPermanent(g, g.Seats[0], testThroneOracle, "Throne") })
	if err := g.Concede(g.Seats[0].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !g.Seats[0].Eliminated {
		t.Error("a concession was replaced")
	}
}

// The planeswalk replacement asks the planeswalking player to order the
// top two cards, and the planeswalk follows the answer.
func TestAPlaneswalkReplacementOrdersTheTopTwoFirst(t *testing.T) {
	withCatalogGameActionReplacements(t)
	g, p := planechaseMain(t)
	g.WithWriteLock(func() { seedReplacementPermanent(g, p, testSusanOracle, "Susan") })
	deck := g.PlanarDeckCards()
	top, second := deck[len(deck)-1].InstanceID, deck[len(deck)-2].InstanceID
	g.WithWriteLock(func() {
		if err := g.PlaneswalkForEffect(p.ID); err != nil {
			t.Fatal(err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d pending choices, want the look-and-order pick", len(g.PendingChoices))
	}
	pc := g.PendingChoices[0]
	if pc.Chooser != p.ID {
		t.Errorf("asked %v, want the planeswalking player", pc.Chooser)
	}
	if err := g.ResolveOptionPick(pc.ID, pc.Chooser, 0); err != nil {
		t.Fatal(err)
	}
	up := g.PlanarFaceUpCards()
	if len(up) != 1 || up[0].InstanceID != second {
		t.Errorf("face up %v, want the card that was second", up)
	}
	if bottom := g.PlanarDeckCards(); len(bottom) == 0 || (bottom[0].InstanceID != top && bottom[1].InstanceID != top) {
		t.Error("the card put on the bottom is not at the bottom")
	}
}
