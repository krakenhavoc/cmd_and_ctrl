package protocol

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_permitted_view_test.go — the wire half of #1573 (ADR 0066's
// 2026-09-24 amendment). A card Gonti, Night Minister or Outrageous
// Robbery exiles face down has ONE viewer, the player who holds the cast
// permission over it: they read the card and get its cast surface; the
// owner and every other seat get a card back labelled `permitted` and
// nothing about casting it.

func TestAFaceDownPermittedCardIsReadableAndCastableOnlyByItsHolder(t *testing.T) {
	g, me, opp := stripTable(t)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	loot := game.NewCard("Wire Stolen Seer", opp.ID)
	loot.TypeLine = "Creature — Eldrazi"
	loot.ManaCost = "{2}{C}"
	loot.OracleID = "test-permitted-seer"
	opp.Library.PushTop(loot)
	g.WithWriteLock(func() {
		if err := g.ExileTopFaceDownWithPermissionForEffect(opp.ID, me.ID, 1, game.CastPermission{
			AnyType:  true,
			Duration: game.WhileInZoneDuration(),
		}); err != nil {
			t.Fatalf("ExileTopFaceDownWithPermissionForEffect: %v", err)
		}
	})
	id := loot.InstanceID

	mine := stripCard(t, g, me.ID.String(), id)
	if mine.Name != "Wire Stolen Seer" || !mine.FaceVisible || !mine.KnownByYou {
		t.Errorf("holder reads name=%q face_visible=%v known=%v — want the card", mine.Name, mine.FaceVisible, mine.KnownByYou)
	}
	if mine.ExilePlay == nil || mine.ExilePlay.Player != me.ID.String() || !mine.ExilePlay.AnyType || !mine.ExilePlay.AnyColor {
		t.Errorf("holder's exile_play = %+v, want theirs with any_type and any_color", mine.ExilePlay)
	}
	if !mine.CastableHere {
		t.Error("the holder may cast it in their main phase")
	}
	// Any type folds the {C} into generic, and the badge says so.
	assertPrice(t, "permitted/holder", mine, "{3}", true, "")

	for _, seat := range []*game.Player{opp, third} {
		label := "permitted/" + seat.Name
		theirs := stripCard(t, g, seat.ID.String(), id)
		if theirs.Name != "" || theirs.ManaCost != "" || theirs.TypeLine != "" || theirs.FaceVisible || theirs.KnownByYou {
			t.Errorf("%s reads name=%q mana_cost=%q type_line=%q face_visible=%v — want a card back",
				label, theirs.Name, theirs.ManaCost, theirs.TypeLine, theirs.FaceVisible)
		}
		if !theirs.FaceDown || theirs.FaceDownKind != string(game.FaceDownPermitted) {
			t.Errorf("%s: face_down=%v kind=%q, want the public kind %q", label, theirs.FaceDown, theirs.FaceDownKind, game.FaceDownPermitted)
		}
		if theirs.ExilePlay != nil {
			t.Errorf("%s got exile_play %+v off a card they cannot read", label, theirs.ExilePlay)
		}
		assertNothingFor(t, label, theirs)
	}
	// The admin / replay frame ("") sees every card by design
	// (FilterViewFor's contract), but it is nobody's seat, so it gets no
	// cast surface.
	assertNothingFor(t, "permitted/unseated", stripCard(t, g, "", id))

	// The log line for the exile names the card to nobody but the holder.
	for _, seat := range []*game.Player{opp, third} {
		for _, e := range ViewOfGameFor(g, seat.ID.String()).Log {
			if strings.Contains(e.Text, "Wire Stolen Seer") {
				t.Errorf("%s's log names the card: %q", seat.Name, e.Text)
			}
		}
	}
}

// An any-type grant over a face-up card: the wire carries the label for
// the holder, and the price is the folded one.
func TestAnAnyTypeGrantIsLabelledOnTheWire(t *testing.T) {
	g, me, opp := stripTable(t)
	id := exileWithGrant(t, g, exiledSpell(opp.ID, "Wire Reality Smasher", "Creature — Eldrazi", "{4}{C}"),
		game.CastPermission{Player: me.ID, AnyType: true, CastOnly: true})
	c := stripCard(t, g, me.ID.String(), id)
	if c.ExilePlay == nil || !c.ExilePlay.AnyType || !c.ExilePlay.AnyColor {
		t.Fatalf("exile_play = %+v, want any_type and any_color", c.ExilePlay)
	}
	assertPrice(t, "any type", c, "{5}", true, "")

	colourOnly := exileWithGrant(t, g, exiledSpell(opp.ID, "Wire Thought-Knot", "Creature — Eldrazi", "{3}{C}"),
		game.CastPermission{Player: me.ID, AnyColor: true})
	if v := stripCard(t, g, me.ID.String(), colourOnly); v.ExilePlay == nil || v.ExilePlay.AnyType {
		t.Errorf("an any-colour grant claims any type: %+v", v.ExilePlay)
	} else {
		assertPrice(t, "any color", v, "{3}{C}", true, "")
	}
}
