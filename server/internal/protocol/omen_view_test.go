package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// omen_view_test.go — the wire half of CR 720.3 (#5175).
//
// An omen card is laid out as an adventure card, and the view offers
// its halves through the same gate the engine and the bot enumerator
// read (game.CastableFaceGrantsLocked), so the Omen half in hand ships
// its own announce block exactly as an Adventure half does.

// TestOmenHalfInHandShipsItsOwnAnnounceBlock: the Omen half's target
// clause rides faces[1], priced off the Omen's own cost (CR 720.3a).
func TestOmenHalfInHandShipsItsOwnAnnounceBlock(t *testing.T) {
	withFaceCatalog(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)

	c := adventureCard(me.ID, me.ID)
	c.Faces[0].TypeLine = "Creature — Dragon"
	c.Faces[1].TypeLine = "Sorcery — Omen"
	c.SetFace(0)
	id := c.InstanceID
	me.Hand.PushTop(c)

	hand := ViewOfGameFor(g, me.ID.String()).Seats[0].Hand
	omen := faceOf(t, hand, id, 1)
	if omen.TypeLine != "Sorcery — Omen" || omen.ManaCost != "{1}{R}" {
		t.Errorf("faces[1] = %q %q, want the Omen's own type line and cost", omen.TypeLine, omen.ManaCost)
	}
	if omen.TargetMode != "any" || omen.LegalTargets == nil {
		t.Fatalf("faces[1] ships no announce block (mode %q, targets %+v): the Omen half has no picker",
			omen.TargetMode, omen.LegalTargets)
	}
	if len(omen.LegalTargets.Cards) != 1 || omen.LegalTargets.Cards[0] != bear.InstanceID.String() {
		t.Errorf("faces[1].legal_targets.cards = %v, want just the bear", omen.LegalTargets.Cards)
	}
	if front := faceOf(t, hand, id, 0); front.TargetMode != "" {
		t.Errorf("faces[0] (the Dragon) carries the Omen's clause: mode %q", front.TargetMode)
	}
}
