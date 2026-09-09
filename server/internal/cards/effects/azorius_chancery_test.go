package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const azoriusChanceryOracle = "189fc8f4-17ac-4f1d-82c8-8401445bdaf4"

// chanceryPlay puts a Chancery into owner's hand and moves it to the
// battlefield, which is the path a played land takes.
func chanceryPlay(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p := playerByIDForTest(g, owner)
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Azorius Chancery", TypeLine: "Land",
		OracleID: azoriusChanceryOracle, Owner: owner, Controller: owner,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		id,
	); err != nil {
		t.Fatalf("play Chancery: %v", err)
	}
	return id
}

func playerByIDForTest(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

// It arrives tapped, so it cannot be used for mana the turn it lands.
func TestAzoriusChanceryEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	aangPushLand(g, p0.ID, "Island", false) // a second land to bounce
	id := chanceryPlay(t, g, p0.ID)

	got, ok := aangCardOnBF(g, id)
	if !ok {
		t.Fatal("Chancery never reached the battlefield")
	}
	if !got.Tapped {
		t.Error("Chancery entered untapped; the enters-tapped clause is its drawback")
	}
}

// The ETB returns the chosen land to its owner's hand.
func TestAzoriusChanceryReturnsChosenLand(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	island := aangPushLand(g, p0.ID, "Island", false)
	handBefore := len(p0.Hand.Cards)

	chancery := chanceryPlay(t, g, p0.ID)
	pickCard(t, g, p0.ID, island)
	passPriorityAroundTable(t, g)

	if _, stillOut := aangCardOnBF(g, island); stillOut {
		t.Error("chosen land is still on the battlefield")
	}
	found := false
	for _, c := range p0.Hand.Cards {
		if c.InstanceID == island {
			found = true
		}
	}
	if !found {
		t.Error("chosen land did not return to its owner's hand")
	}
	if len(p0.Hand.Cards) != handBefore+1 {
		t.Errorf("hand %d, want %d", len(p0.Hand.Cards), handBefore+1)
	}
	// The Chancery itself stays put when something else was chosen.
	if _, ok := aangCardOnBF(g, chancery); !ok {
		t.Error("Chancery left the battlefield despite another land being chosen")
	}
}

// With no other land, the Chancery is still a legal choice for its own
// trigger — so the trigger can never fizzle for want of a target.
func TestAzoriusChanceryCanReturnItself(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	chancery := chanceryPlay(t, g, p0.ID)

	pickCard(t, g, p0.ID, chancery)
	passPriorityAroundTable(t, g)

	if _, stillOut := aangCardOnBF(g, chancery); stillOut {
		t.Error("Chancery did not return itself")
	}
	found := false
	for _, c := range p0.Hand.Cards {
		if c.InstanceID == chancery {
			found = true
		}
	}
	if !found {
		t.Error("Chancery did not end up in its owner's hand")
	}
}

func TestAzoriusChanceryManaAbility(t *testing.T) {
	ab := game.ManaAbilitiesForCard(game.Card{OracleID: azoriusChanceryOracle})
	if len(ab) != 1 {
		t.Fatalf("%d mana abilities, want 1", len(ab))
	}
	if ab[0].Produced != "{W}{U}" {
		t.Errorf("produced %q, want {W}{U}", ab[0].Produced)
	}
}
