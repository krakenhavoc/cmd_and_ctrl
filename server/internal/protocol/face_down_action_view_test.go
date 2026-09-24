package protocol

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_action_view_test.go — ADR 0082 decision 9: the
// turn-face-up row on the wire.
//
// The face-down permanent's BODY and its redaction are ADR 0069's and
// are tested in face_down_view_test.go. What is new here is one
// projection: `special_actions`, which used to reach only a card in
// the viewer's own hand and now reaches a face-down permanent too.

const morphViewOracle = "oracle-morph-view"

// withMorphOffer wires a catalog morph declaration for one test.
func withMorphOffer(t *testing.T, faceUp string) {
	t.Helper()
	prev := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(oracleID string) []game.AlternativeCost {
		if oracleID != morphViewOracle {
			return nil
		}
		return []game.AlternativeCost{{
			Key:      "morph",
			Label:    "Morph — cast face down {3}",
			ManaCost: "{3}",
			FaceDown: &game.FaceDownCast{Kind: game.FaceDownMorphed, FaceUpCost: faceUp},
		}}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prev })
}

// TestTurnFaceUpRowReachesTheControllerAlone: a face-down permanent
// carries its CR 116.2g "turn face up" row, on the battlefield, for
// the one player CR 708.5 lets look at it.
//
// The row quotes the morph cost, which names the card as loudly as a
// mana cost does — so it has to be stripped for everybody else, and it
// is, by the same redaction that strips the hand's foretell rows.
func TestTurnFaceUpRowReachesTheControllerAlone(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	withMorphOffer(t, "{1}{U}")

	c := game.NewCard("Willbender", me.ID)
	c.OracleID = morphViewOracle
	c.TypeLine = "Creature — Human Wizard"
	c.ManaCost = "{1}{U}"
	id := c.InstanceID
	g.WithWriteLock(func() {
		me.Library.PushTop(c)
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
		// The manifest is how a test makes a face-down permanent
		// without a cast. What is under test is the ROW, and the
		// MORPHED state is the one that prices it off the card's own
		// morph cost rather than off its mana cost (CR 708.6 vs
		// CR 701.40b) — which is also the arm that proves the
		// projection reads the engine's per-kind rule rather than
		// guessing.
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SetFaceDown(game.FaceDownMorphed)
			}
		}
	})

	find := func(viewer uuid.UUID) CardView {
		t.Helper()
		for _, cv := range ViewOfGameFor(g, viewer.String()).Battlefield.Cards {
			if cv.InstanceID == id.String() {
				return cv
			}
		}
		t.Fatalf("face-down permanent missing from %s's battlefield view", viewer)
		return CardView{}
	}

	mine := find(me.ID)
	if len(mine.SpecialActions) != 1 {
		t.Fatalf("controller: special_actions = %+v, want one turn_face_up row", mine.SpecialActions)
	}
	row := mine.SpecialActions[0]
	if row.Kind != "turn_face_up" {
		t.Errorf("controller: kind = %q, want turn_face_up", row.Kind)
	}
	if row.Cost != "{1}{U}" {
		t.Errorf("controller: cost = %q, want the card's morph cost (CR 708.6)", row.Cost)
	}
	if !row.Available {
		t.Error("controller: available = false; CR 702.37c is any time you have priority")
	}

	if got := find(opp.ID).SpecialActions; len(got) != 0 {
		t.Errorf("opponent: special_actions = %+v, want none — the row quotes the card", got)
	}
}

// TestAFaceUpPermanentCarriesNoSpecialActionRow: the offer is DERIVED
// from the face-down kind, so stamping the projection onto every
// battlefield permanent adds a row to none of them.
func TestAFaceUpPermanentCarriesNoSpecialActionRow(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	withMorphOffer(t, "{1}{U}")

	c := game.NewCard("Willbender", me.ID)
	c.OracleID = morphViewOracle
	c.TypeLine = "Creature — Human Wizard"
	c.ManaCost = "{1}{U}"
	c.Controller = me.ID
	id := c.InstanceID
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })

	for _, cv := range ViewOfGameFor(g, me.ID.String()).Battlefield.Cards {
		if cv.InstanceID != id.String() {
			continue
		}
		if len(cv.SpecialActions) != 0 {
			t.Errorf("special_actions = %+v, want none: a face-UP permanent has no way to be turned up", cv.SpecialActions)
		}
		return
	}
	t.Fatal("the permanent is missing from its controller's view")
}
