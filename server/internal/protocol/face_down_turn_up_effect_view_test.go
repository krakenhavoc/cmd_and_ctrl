package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_turn_up_effect_view_test.go — #2590, ADR 0082's second
// 2026-10-07 amendment: the log's half of turning a permanent face up
// as an effect. The CR 116.2g special action narrates itself through
// LogSpecialAction; only an effect's turn needs a line of its own.

func turnUpLines(g *game.Game, viewer uuid.UUID, id uuid.UUID) []LogEvent {
	var out []LogEvent
	for _, e := range ViewOfGameFor(g, viewer.String()).Log {
		if e.Kind == LogTurnFaceUp && e.CardID == id.String() {
			out = append(out, e)
		}
	}
	return out
}

func TestAnEffectsTurnFaceUpIsNarratedAndNamesBothSides(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := turnedPermanent(t, g, me, "Hauntwoods Shrieker", "oracle-shrieker")
	id := turnedPermanent(t, g, opp, "Sheoldred, the Apocalypse", "oracle-sheoldred")

	var turned bool
	g.WithWriteLock(func() { turned = g.TurnFaceUpForEffect(src, me.ID, id) })
	if !turned {
		t.Fatal("the effect door refused")
	}

	for _, seat := range []*game.Player{me, opp} {
		lines := turnUpLines(g, seat.ID, id)
		if len(lines) != 1 {
			t.Fatalf("%s: %d turn_face_up lines, want 1", seat.ID, len(lines))
		}
		e := lines[0]
		if !strings.Contains(e.Text, "face up") {
			t.Errorf("%s: %q does not say a permanent was turned face up", seat.ID, e.Text)
		}
		if !strings.Contains(e.Text, "Sheoldred") {
			t.Errorf("%s: %q does not name the permanent, which is public again", seat.ID, e.Text)
		}
		if e.Target != src.String() {
			t.Errorf("%s: target = %q, want the source %q", seat.ID, e.Target, src)
		}
	}
}

func TestTheSpecialActionsTurnFaceUpHasNoLineOfItsOwn(t *testing.T) {
	withMorphOffer(t, "{1}{U}")
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := turnedPermanent(t, g, me, "Willbender", morphViewOracle)
	g.WithWriteLock(func() {
		for i := 0; i < 8; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "U"})
		}
		g.Turn.Step = game.StepPrecombatMain
	})
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("special action: %v", err)
	}
	if lines := turnUpLines(g, me.ID, id); len(lines) != 0 {
		t.Errorf("the special action wrote %d turn_face_up lines; its LogSpecialAction line already says it", len(lines))
	}
}
