package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A discarded creature that is exiled before the {2}{U} is answered is
// still copied: the copy is made from the card's copiable values (CR 707.2),
// which do not depend on the zone it sits in, as the last-known-information
// reading of the printed text also gives (CR 608.2h). The old caveat read
// this as a gap; it is the printed result.
func TestHashatonStillCopiesADiscardedCreatureExiledInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ogre := hashatonDiscardOgre(t, g, me)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("discarding a creature did not offer the {2}{U} payment")
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile},
		ogre,
	); err != nil {
		t.Fatalf("exile the discarded card: %v", err)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	answerPayUnless(t, g, me.ID, true)

	var tok uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Test Ogre" && c.InstanceID != ogre {
			tok = c.InstanceID
		}
	}
	if tok == uuid.Nil {
		t.Fatal("the exiled card was not copied")
	}
	c, _ := battlefieldCard(g, tok)
	if c.Power != 4 || c.Toughness != 4 || !c.Tapped {
		t.Errorf("token is %d/%d tapped=%v, want a tapped 4/4", c.Power, c.Toughness, c.Tapped)
	}
}
