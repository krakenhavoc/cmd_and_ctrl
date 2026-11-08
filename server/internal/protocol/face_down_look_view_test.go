package protocol

import (
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	// The real Found Footage catalog row.
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_look_view_test.go — #5251: the wire's half of "you may look
// at face-down creatures you don't control any time". The grant makes
// its holder's seat, and only that seat, a knower of each face-down
// creature it covers: never a spectator (who knows only what every seat
// knows, ADR 0069's 2026-09-30 amendment), never another opponent,
// never a replay, and never once the grant has ended.

const foundFootageViewOracle = "ad37551c-9d20-4902-8190-6a0aa32a7947"

// pushFoundFootage puts a Found Footage onto the battlefield under
// `controller` and returns its ID.
func pushFoundFootage(g *game.Game, controller uuid.UUID) uuid.UUID {
	c := game.NewCard("Found Footage", controller)
	c.Controller = controller
	c.OracleID = foundFootageViewOracle
	c.TypeLine = "Artifact — Clue"
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

func lookViewOf(t *testing.T, g *game.Game, viewer string, id uuid.UUID) CardView {
	t.Helper()
	for _, cv := range ViewOfGameFor(g, viewer).Battlefield.Cards {
		if cv.InstanceID == id.String() {
			return cv
		}
	}
	t.Fatalf("permanent %s missing from %q's battlefield view", id, viewer)
	return CardView{}
}

// lookFaceDown seats a public face-up Sheoldred under `controller` and
// turns it face down through the engine's own door, so its KnownBy is a
// real face-down permanent's: the controller alone. The art is what a
// viewer who may look gets (the object itself has no name, CR 708.2a).
func lookFaceDown(t *testing.T, g *game.Game, controller *game.Player) uuid.UUID {
	t.Helper()
	c := game.NewCard("Sheoldred, the Apocalypse", controller.ID)
	c.Controller = controller.ID
	c.OracleID = "oracle-look-view"
	c.ScryfallID = lookScryfall
	c.TypeLine = "Legendary Creature — Phyrexian Praetor"
	c.ManaCost = "{2}{B}{B}"
	c.Power, c.Toughness = 4, 5
	c.EnteredBattlefieldAt = time.Now().UnixNano()
	for _, p := range g.Seats {
		c.AddKnower(p.ID)
	}
	id := c.InstanceID
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(c)
		if len(g.TurnFaceDownForEffect(uuid.Nil, id)) != 1 {
			t.Fatal("the creature was not turned face down")
		}
	})
	return id
}

const lookScryfall = "scryfall-sheoldred"

func assertLookSees(t *testing.T, label string, c CardView) {
	t.Helper()
	if !c.KnownByYou || !c.FaceVisible || c.ScryfallID != lookScryfall {
		t.Errorf("%s: known %v, face visible %v, art %q; want the card", label, c.KnownByYou, c.FaceVisible, c.ScryfallID)
	}
}

// assertLookHidden is the redaction allowlist plus the public CR 708.2
// body (type line "Creature", 2/2), which every seat sees.
func assertLookHidden(t *testing.T, label string, c CardView) {
	t.Helper()
	if c.KnownByYou || c.FaceVisible || c.Name != "" || c.ScryfallID != "" {
		t.Errorf("%s: known %v, face visible %v, name %q, art %q; want a face-down card", label, c.KnownByYou, c.FaceVisible, c.Name, c.ScryfallID)
	}
	if !c.FaceDown || c.TypeLine != "Creature" || c.Power != 2 || c.Toughness != 2 {
		t.Errorf("%s: face down %v, %q %d/%d; want a face-down 2/2", label, c.FaceDown, c.TypeLine, c.Power, c.Toughness)
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatal(err)
	}
	var leaked []string
	for k := range keys {
		if !redactedCardKeys[k] && k != "type_line" && k != "power" && k != "toughness" {
			leaked = append(leaked, k)
		}
	}
	sort.Strings(leaked)
	if len(leaked) > 0 {
		t.Errorf("%s: a viewer who may not look received %v\n%s", label, leaked, raw)
	}
}

// Three seats: the holder sees the card, the creature's controller still
// does, and another opponent and a spectator see a face-down 2/2. The
// source leaving hides it again on the next frame.
func TestALookGrantShowsTheCardToItsHolderAlone(t *testing.T) {
	g := buildThreeSeatGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	id := lookFaceDown(t, g, opp)
	footage := pushFoundFootage(g, me.ID)

	assertLookSees(t, "the holder", lookViewOf(t, g, me.ID.String(), id))
	assertLookSees(t, "the controller", lookViewOf(t, g, opp.ID.String(), id))
	assertLookHidden(t, "another opponent", lookViewOf(t, g, other.ID.String(), id))
	assertLookHidden(t, "a spectator", lookViewOf(t, g, SpectatorViewerID, id))
	if admin := lookViewOf(t, g, "", id); !admin.KnownByYou {
		t.Error("the admin's omniscient view lost the card")
	}

	g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(footage) })
	assertLookHidden(t, "the former holder, after the source left", lookViewOf(t, g, me.ID.String(), id))
}

// Two seats: the holder and the controller are EVERY seat, which is the
// case a knower-based grant would leak to the spectator through
// knownToEverySeat. The grant is not a reveal to the table.
func TestALookGrantInATwoPlayerGameHidesTheCardFromASpectator(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := lookFaceDown(t, g, opp)
	pushFoundFootage(g, me.ID)

	assertLookSees(t, "the holder", lookViewOf(t, g, me.ID.String(), id))
	assertLookHidden(t, "a spectator", lookViewOf(t, g, SpectatorViewerID, id))
}

// The until-end-of-turn record reaches the view the same way.
func TestAnUntilEndOfTurnLookGrantShowsTheCardToItsHolderAlone(t *testing.T) {
	g := buildThreeSeatGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	id := lookFaceDown(t, g, opp)
	g.WithWriteLock(func() {
		if !g.LookAtFaceDownCreaturesThisTurnForEffect(uuid.Nil, other.ID, "Lumbering Laundry") {
			t.Fatal("no record was written")
		}
	})
	assertLookSees(t, "the holder", lookViewOf(t, g, other.ID.String(), id))
	assertLookHidden(t, "another opponent", lookViewOf(t, g, me.ID.String(), id))
	assertLookHidden(t, "a spectator", lookViewOf(t, g, SpectatorViewerID, id))
}

// A replay is the unfiltered frame written as JSON and filtered again
// when it is downloaded (lobby.streamFilteredReplay). The grant is never
// on the wire, so the holder's replay shows a face-down 2/2, as the
// spectator's does.
func TestALookGrantIsNotInAReplayFrame(t *testing.T) {
	g := buildThreeSeatGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := lookFaceDown(t, g, opp)
	pushFoundFootage(g, me.ID)

	raw, err := json.Marshal(ViewOfGame(g))
	if err != nil {
		t.Fatal(err)
	}
	var frame GameView
	if err := json.Unmarshal(raw, &frame); err != nil {
		t.Fatal(err)
	}
	for _, viewer := range []string{me.ID.String(), SpectatorViewerID} {
		for _, cv := range FilterViewFor(frame, viewer).Battlefield.Cards {
			if cv.InstanceID == id.String() {
				assertLookHidden(t, "replay for "+viewer, cv)
			}
		}
	}
}
