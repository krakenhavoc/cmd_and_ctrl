package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_choice_test.go — #984. A colour, a creature type and a player
// chosen for a permanent are all announced at the table (CR 105.4,
// CR 614.12) and are all read back by the card's own later abilities
// (CR 607.2d). #781 put them on the CARD; these are the lines that say
// WHO chose WHAT, and WHEN.
//
// The one thing every assertion below shares: the text names people
// and cards by name and never carries a UUID. A log line with an
// instance ID in it is unreadable to a player and, for a card in a
// hidden zone, is the correlation handle the whole file is built to
// keep off the wire.

// assertNoUUID fails when a rendered line carries anything shaped like
// an instance ID. Cheap and blunt on purpose: the shape is what a
// reader notices, and a line that has one is wrong whichever ID it is.
func assertNoUUID(t *testing.T, text string) {
	t.Helper()
	for _, field := range strings.Fields(text) {
		trimmed := strings.Trim(field, ".,:;()\"")
		if _, err := uuid.Parse(trimmed); err == nil {
			t.Errorf("log line carries a UUID: %q", text)
			return
		}
	}
}

func TestLogNarratesAChosenColor(t *testing.T) {
	g := buildActiveGame(t)
	chooser, other := g.Seats[0], g.Seats[1]
	heart := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: heart, Name: "Coldsteel Heart", TypeLine: "Artifact",
			Owner: chooser.ID, Controller: chooser.ID,
			ChosenColor: "G",
			KnownBy:     map[uuid.UUID]bool{chooser.ID: true, other.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventColorChosen, Actor: chooser.ID, CardID: heart, Label: "G",
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogChooseColor)
	if want := "P1 chose green for Coldsteel Heart"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Choice != "G" {
		t.Errorf("choice: got %q, want %q", entry.Choice, "G")
	}
	if entry.CardID != heart.String() {
		t.Errorf("card_id: got %q, want the heart", entry.CardID)
	}
	if entry.Seat != 0 {
		t.Errorf("seat: got %d, want 0", entry.Seat)
	}
	assertNoUUID(t, entry.Text)
}

func TestLogNarratesAChosenCreatureType(t *testing.T) {
	g := buildActiveGame(t)
	chooser, other := g.Seats[0], g.Seats[1]
	cavern := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: cavern, Name: "Cavern of Souls", TypeLine: "Land",
			Owner: chooser.ID, Controller: chooser.ID,
			NamedTribe: "Elf",
			KnownBy:    map[uuid.UUID]bool{chooser.ID: true, other.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCreatureTypeChosen, Actor: chooser.ID, CardID: cavern, Label: "Elf",
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogChooseType)
	if want := "P1 chose Elf for Cavern of Souls"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Choice != "Elf" {
		t.Errorf("choice: got %q, want %q", entry.Choice, "Elf")
	}
	assertNoUUID(t, entry.Text)
}

// TestLogNarratesAChosenPlayer is the #1007 half of #984: the third
// as-enters choice shipped with the same hole the other two had.
//
// The answer is a SEAT INDEX, not a name in `choice` and not a UUID in
// `target` — the same convention every other player reference in the
// log follows, so a client that highlights seats highlights this one
// for free.
func TestLogNarratesAChosenPlayer(t *testing.T) {
	g := buildActiveGame(t)
	chooser, chosen := g.Seats[0], g.Seats[1]
	nemesis := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: nemesis, Name: "True-Name Nemesis", TypeLine: "Creature — Merfolk Rogue",
			Owner: chooser.ID, Controller: chooser.ID,
			ChosenPlayer: chosen.ID,
			KnownBy:      map[uuid.UUID]bool{chooser.ID: true, chosen.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventPlayerChosen, Actor: chooser.ID, CardID: nemesis,
			Target: chosen.ID, Label: "P2",
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogChoosePlayer)
	if want := "P1 chose P2 for True-Name Nemesis"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.TargetSeat == nil || *entry.TargetSeat != 1 {
		t.Errorf("target_seat: got %v, want seat 1", entry.TargetSeat)
	}
	if entry.Target != "" {
		t.Errorf("target: got %q, want empty — a chosen player is a seat, never a card", entry.Target)
	}
	if entry.Choice != "" {
		t.Errorf("choice: got %q, want empty — the seat index already says it", entry.Choice)
	}
	assertNoUUID(t, entry.Text)
}

// TestChosenPlayerOfADepartedSeatNamesNoUUID is the fail-closed half.
// A seat the view no longer carries must not fall through onto
// `target` as if it were a card instance.
func TestChosenPlayerOfADepartedSeatNamesNoUUID(t *testing.T) {
	g := buildActiveGame(t)
	chooser := g.Seats[0]
	nemesis, ghost := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: nemesis, Name: "True-Name Nemesis", TypeLine: "Creature — Merfolk Rogue",
			Owner: chooser.ID, Controller: chooser.ID,
			KnownBy: map[uuid.UUID]bool{chooser.ID: true, g.Seats[1].ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventPlayerChosen, Actor: chooser.ID, CardID: nemesis, Target: ghost,
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogChoosePlayer)
	if want := "P1 chose a player for True-Name Nemesis"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Target != "" || entry.TargetSeat != nil {
		t.Errorf("an unseatable answer leaked: target %q target_seat %v", entry.Target, entry.TargetSeat)
	}
	assertNoUUID(t, entry.Text)
}

// TestAChosenValueIsRedactedWithTheCardThatAsked — CR 105.4 makes the
// answer public, but the answer also IDENTIFIES the card ("Elf" names
// Cavern of Souls), which is why #781 strips ChosenColor / NamedTribe
// from a CardView the viewer is not a knower of. The log has to make
// the same trade or it is the leak the card view closed.
//
// Nothing in the catalog can produce this today — a face-down
// permanent has no abilities, so it asks nothing — so this is the log
// failing closed ahead of the engine, the way revealEntry does.
func TestAChosenValueIsRedactedWithTheCardThatAsked(t *testing.T) {
	g := buildActiveGame(t)
	chooser, other := g.Seats[0], g.Seats[1]
	secret := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: secret, Name: "Hidden Chooser", TypeLine: "Artifact",
			Owner: chooser.ID, Controller: chooser.ID,
			KnownBy: map[uuid.UUID]bool{chooser.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventColorChosen, Actor: chooser.ID, CardID: secret, Label: "G",
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCreatureTypeChosen, Actor: chooser.ID, CardID: secret, Label: "Elf",
		})
	})

	mine := FilterViewFor(ViewOfGame(g), chooser.ID.String())
	colorMine := findLog(t, mine.Log, LogChooseColor)
	if want := "P1 chose green for Hidden Chooser"; colorMine.Text != want {
		t.Errorf("the chooser's own line: got %q, want %q", colorMine.Text, want)
	}

	theirs := FilterViewFor(ViewOfGame(g), other.ID.String())
	colorTheirs := findLog(t, theirs.Log, LogChooseColor)
	if want := "P1 chose a color for a card"; colorTheirs.Text != want {
		t.Errorf("a non-knower's line: got %q, want %q", colorTheirs.Text, want)
	}
	if colorTheirs.Choice != "" {
		t.Errorf("the colour survived the redaction as %q", colorTheirs.Choice)
	}
	typeTheirs := findLog(t, theirs.Log, LogChooseType)
	if want := "P1 chose a creature type for a card"; typeTheirs.Text != want {
		t.Errorf("a non-knower's type line: got %q, want %q", typeTheirs.Text, want)
	}
	if typeTheirs.Choice != "" {
		t.Errorf("the creature type survived the redaction as %q", typeTheirs.Choice)
	}
}

// TestAChosenColorForAVanishedCardStillNamesTheColour separates the
// two "no card name" cases: a card the view cannot account for was
// never REDACTED, so the public answer stays on the line.
func TestAChosenColorForAVanishedCardStillNamesTheColour(t *testing.T) {
	g := buildActiveGame(t)
	chooser := g.Seats[0]
	gone := uuid.New()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventColorChosen, Actor: chooser.ID, CardID: gone, Label: "R",
		})
	})

	entry := findLog(t, FilterViewFor(ViewOfGame(g), chooser.ID.String()).Log, LogChooseColor)
	if want := "P1 chose red for a card"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Choice != "R" {
		t.Errorf("choice: got %q, want %q", entry.Choice, "R")
	}
	assertNoUUID(t, entry.Text)
}
