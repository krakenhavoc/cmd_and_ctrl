package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pull_from_eternity_test.go -- #3259. The first catalog card on a clause
// over cards in exile (TargetCardInExile).

const pullFromEternityOracle = "2903bd66-6948-445b-a18a-12e2ce1a86c3"

func pushExiledCard(g *game.Game, owner uuid.UUID, name, typeLine string, faceDown bool) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner, FaceDown: faceDown,
	})
	return id
}

// The spell puts the targeted face-up exiled card into its OWNER's
// graveyard, whoever cast the spell.
func TestPullFromEternityPutsAnExiledCardIntoItsOwnersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	exiled := pushExiledCard(g, opp.ID, "Exiled Land", "Land", false)

	castCatalogSpell(t, g, "Pull from Eternity", "Instant", pullFromEternityOracle, cardRefs(exiled))
	passPriorityAroundTable(t, g)

	if g.Exile.Contains(exiled) {
		t.Error("the card is still in exile")
	}
	if !opp.Graveyard.Contains(exiled) {
		t.Error("the card did not land in its owner's graveyard")
	}
}

// A face-down exiled card is not a "face-up exiled card": the announce
// refuses it even though it is in exile.
func TestPullFromEternityRefusesAFaceDownCard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	hidden := pushExiledCard(g, opp.ID, "Foretold Card", "Sorcery", true)
	// Known to the caster -- it is still face down.
	g.WithWriteLock(func() {
		for i := range g.Exile.Cards {
			if g.Exile.Cards[i].InstanceID == hidden {
				g.Exile.Cards[i].AddKnower(g.Seats[0].ID)
			}
		}
	})
	err := castCatalogSpellErr(t, g, "Pull from Eternity", "Instant", pullFromEternityOracle, cardRefs(hidden))
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a face-down exiled card: err = %v, want ErrIllegalTarget", err)
	}
}

// A card on the battlefield or in a graveyard is not in exile.
func TestPullFromEternityRefusesACardOutsideExile(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	err := castCatalogSpellErr(t, g, "Pull from Eternity", "Instant", pullFromEternityOracle, cardRefs(bear))
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a battlefield card: err = %v, want ErrIllegalTarget", err)
	}
}

// The target is re-checked as the spell resolves (CR 608.2b): a card
// that left exile in response is gone, and the spell does nothing to it.
func TestPullFromEternityIsReCheckedOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	exiled := pushExiledCard(g, opp.ID, "Exiled Land", "Land", false)

	castCatalogSpell(t, g, "Pull from Eternity", "Instant", pullFromEternityOracle, cardRefs(exiled))
	// In response the card returns to its owner's hand.
	g.WithWriteLock(func() {
		c, err := g.Exile.Remove(exiled)
		if err != nil {
			t.Fatalf("Remove: %v", err)
		}
		opp.Hand.PushTop(c)
	})
	passPriorityAroundTable(t, g)

	if opp.Graveyard.Contains(exiled) {
		t.Error("the spell reached a card that had left exile")
	}
	if !opp.Hand.Contains(exiled) {
		t.Error("the card should still be where the response put it")
	}
}
