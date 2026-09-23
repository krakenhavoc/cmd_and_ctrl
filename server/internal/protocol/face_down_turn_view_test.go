package protocol

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_turn_view_test.go — #1209, ADR 0082's 2026-09-23
// amendment: the wire's half of turning a permanent face down.
//
// Two projections, and both are about the same fact — the permanent
// was PUBLIC a moment ago and is not any more (CR 708.5):
//
//  1. the CARD, which stops naming itself to everyone but its
//     controller and starts shipping the CR 708.2 body instead;
//  2. the LOG line, which is the one place that narrowing could leak
//     back out, because the history is rendered once and read by the
//     whole table.
//
// The turn_face_up ROW is the third, and it is the interesting one:
// an Ixidron'd vanilla gets none (CR 708.7) and a Backslid morph gets
// one (CR 702.37e), from the same projection with no view-side rule.

// turnedPermanent seats a public face-up creature and then turns it
// face down through the engine's own primitive, so the view tests are
// reading the state a real Ixidron produces rather than one a test
// stamped by hand.
func turnedPermanent(t *testing.T, g *game.Game, owner *game.Player, name, oracle string) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.Controller = owner.ID
	c.OracleID = oracle
	c.TypeLine = "Creature — Phyrexian Praetor"
	c.ManaCost = "{2}{B}{B}"
	c.Colors = []string{"B"}
	c.Power, c.Toughness = 4, 5
	c.EnteredBattlefieldAt = time.Now().UnixNano()
	c.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		c.KnownBy[p.ID] = true
	}
	id := c.InstanceID
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(c)
		g.TurnFaceDownForEffect(uuid.Nil, id)
	})
	return id
}

func findBattlefieldView(t *testing.T, g *game.Game, viewer uuid.UUID, id uuid.UUID) CardView {
	t.Helper()
	for _, cv := range ViewOfGameFor(g, viewer.String()).Battlefield.Cards {
		if cv.InstanceID == id.String() {
			return cv
		}
	}
	t.Fatalf("permanent %s missing from %s's battlefield view", id, viewer)
	return CardView{}
}

// TestATurnedPermanentShowsTheCR7082BodyToEveryoneButItsController is
// ADR 0069's redaction reached by the new kind. Nothing in view.go
// knows about #1209; this is the assertion that it does not have to.
func TestATurnedPermanentShowsTheCR7082BodyToEveryoneButItsController(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := turnedPermanent(t, g, me, "Sheoldred, the Apocalypse", "oracle-turned-view")

	theirs := findBattlefieldView(t, g, opp.ID, id)
	if theirs.KnownByYou {
		t.Error("opponent: may look at a permanent that was turned face down (CR 708.5)")
	}
	if !theirs.FaceDown {
		t.Error("opponent: the permanent does not read as face down")
	}
	if theirs.FaceDownKind != string(game.FaceDownTurned) {
		t.Errorf("opponent: face_down_kind = %q, want %q", theirs.FaceDownKind, game.FaceDownTurned)
	}
	if theirs.Name != "" || theirs.Power != 2 || theirs.Toughness != 2 {
		t.Errorf("opponent: %q %d/%d, want the public CR 708.2 body: no name, 2/2",
			theirs.Name, theirs.Power, theirs.Toughness)
	}

	mine := findBattlefieldView(t, g, me.ID, id)
	if !mine.KnownByYou {
		t.Error("controller: may not look at their own face-down permanent (CR 708.5)")
	}
	if mine.Power != 2 || mine.Toughness != 2 {
		t.Errorf("controller: sees a %d/%d, not the 2/2 every seat sees", mine.Power, mine.Toughness)
	}
}

// TestTurnFaceDownLogSaysWhatHappenedAndNamesNobody. The line exists
// because nothing else says a permanent turned over — it is not a
// zone change and not a transform — and it names the SOURCE, which is
// public, and not the permanent, which has stopped having a name for
// anyone at all (CR 708.2a). Both seats read the same sentence, and
// the one that could leak an identity is the one that does not exist.
func TestTurnFaceDownLogSaysWhatHappenedAndNamesNobody(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := game.NewCard("Ixidron", me.ID)
	src.Controller = me.ID
	src.TypeLine = "Creature — Illusion"
	src.EnteredBattlefieldAt = time.Now().UnixNano()
	src.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	victim := game.NewCard("Sheoldred, the Apocalypse", me.ID)
	victim.Controller = me.ID
	victim.TypeLine = "Legendary Creature — Phyrexian Praetor"
	victim.Power, victim.Toughness = 4, 5
	victim.EnteredBattlefieldAt = time.Now().UnixNano()
	victim.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	id := victim.InstanceID
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(src)
		g.Battlefield.PushTop(victim)
		g.TurnFaceDownForEffect(src.InstanceID, id)
	})

	entry := func(viewer uuid.UUID) LogEvent {
		t.Helper()
		for _, e := range ViewOfGameFor(g, viewer.String()).Log {
			if e.Kind == LogTurnFaceDown && e.CardID == id.String() {
				return e
			}
		}
		t.Fatalf("no turn_face_down log entry for %s", viewer)
		return LogEvent{}
	}

	for _, seat := range []*game.Player{me, opp} {
		e := entry(seat.ID)
		if !strings.Contains(e.Text, "face down") {
			t.Errorf("%s: %q must say that something was turned face down — a permanent "+
				"silently becoming a 2/2 is the gap this line exists to close", seat.ID, e.Text)
		}
		if !strings.Contains(e.Text, "Ixidron") {
			t.Errorf("%s: %q does not name the source, which is the public half", seat.ID, e.Text)
		}
		if strings.Contains(e.Text, "Sheoldred") {
			t.Errorf("%s: %q names a permanent that has no name any more (CR 708.2a)", seat.ID, e.Text)
		}
		if e.Target != src.InstanceID.String() {
			t.Errorf("%s: target = %q, want the source %q", seat.ID, e.Target, src.InstanceID)
		}
	}
}

// TestTurnFaceUpRowFollowsTheCardUnderneathNotTheEffect is CR 708.7
// and CR 702.37e on the wire, side by side. Both permanents were
// turned face down by the same primitive; the only difference is
// whether the CARD prints morph, and the projection reads the
// engine's answer rather than deriving one.
func TestTurnFaceUpRowFollowsTheCardUnderneathNotTheEffect(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	withMorphOffer(t, "{1}{U}")

	vanilla := turnedPermanent(t, g, me, "Sheoldred, the Apocalypse", "oracle-no-morph")
	morph := turnedPermanent(t, g, me, "Willbender", morphViewOracle)

	if got := findBattlefieldView(t, g, me.ID, vanilla).SpecialActions; len(got) != 0 {
		t.Errorf("a card with no morph offers %+v — CR 708.7 gives it no way back up", got)
	}
	rows := findBattlefieldView(t, g, me.ID, morph).SpecialActions
	if len(rows) != 1 || rows[0].Kind != "turn_face_up" {
		t.Fatalf("a card with morph offers %+v, want one turn_face_up row (CR 702.37e)", rows)
	}
	if rows[0].Cost != "{1}{U}" {
		t.Errorf("cost = %q, want the card's morph cost", rows[0].Cost)
	}
}
