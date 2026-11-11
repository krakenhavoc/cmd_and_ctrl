package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// format_view_test.go: a format on the wire (ADR 0222 §1 and slice 1).
// A Standard table's view names its format and says no seat has a
// commander; a Commander table's view carries neither key, so its wire
// is what it was before formats.

func formatTestGame(t *testing.T, format game.Format, deck func() []game.Card) *game.Game {
	t.Helper()
	g := game.NewGame()
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{Format: &format}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck()); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	return g
}

func sixtyCards() []game.Card {
	deck := make([]game.Card, 0, 60)
	for i := 0; i < 60; i++ {
		deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", i), uuid.Nil))
	}
	return deck
}

func hundredWithACommander() []game.Card {
	deck := []game.Card{game.NewCommander("Commander", uuid.Nil)}
	for i := 0; i < 99; i++ {
		deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", i), uuid.Nil))
	}
	return deck
}

func TestStandardTableViewNamesItsFormatAndNoCommanders(t *testing.T) {
	g := formatTestGame(t, game.FormatStandard, sixtyCards)
	v := ViewOfGame(g)
	if v.Settings == nil || v.Settings.Format != "standard" || v.Settings.StartingLife != 20 {
		t.Fatalf("settings = %+v, want format standard at 20 life", v.Settings)
	}
	if !v.NoCommanders {
		t.Error("no seat has a commander, but the view does not say so")
	}
	// The same for every viewer, a seat and a spectator alike.
	for _, viewer := range []string{v.Seats[0].ID, ""} {
		f := FilterViewFor(v, viewer)
		if !f.NoCommanders || f.Settings == nil || f.Settings.Format != "standard" {
			t.Errorf("viewer %q: no_commanders %v, settings %+v", viewer, f.NoCommanders, f.Settings)
		}
	}
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"no_commanders":true`, `"format":"standard"`} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("the wire lacks %s", key)
		}
	}
	// The HTTP projection agrees with the view's.
	if got := ViewOfTableSettings(g.TableSettingsSnapshot()); got != *v.Settings {
		t.Errorf("ViewOfTableSettings = %+v, the view has %+v", got, *v.Settings)
	}
}

func TestCommanderTableViewIsUnchangedByFormats(t *testing.T) {
	g := formatTestGame(t, game.FormatCommander, hundredWithACommander)
	v := ViewOfGame(g)
	if v.NoCommanders {
		t.Error("a Commander table's view says no seat has a commander")
	}
	if v.Settings == nil || v.Settings.Format != "" {
		t.Fatalf("settings = %+v, want no format named", v.Settings)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"no_commanders"`, `"format"`, `"max_seats"`} {
		if strings.Contains(string(raw), key) {
			t.Errorf("a Commander table's wire carries %s", key)
		}
	}
	// A commander that has left the command zone is still a commander:
	// the flag reads the cards wherever they are.
	seat := g.Seats[0]
	cmd := seat.Command.Cards[0]
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneCommand, Owner: seat.ID},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: seat.ID}, cmd.InstanceID); err == nil {
		if ViewOfGame(g).NoCommanders {
			t.Error("with a commander in a graveyard the view says no seat has one")
		}
	}

	// A sixty-card table that names no format is still a Commander
	// table by its settings; only the cards hide the commander UI.
	plain := formatTestGame(t, game.FormatCommander, sixtyCards)
	pv := ViewOfGame(plain)
	if !pv.NoCommanders || pv.Settings.Format != "" || pv.Settings.StartingLife != game.StartingLife {
		t.Errorf("no_commanders %v, settings %+v", pv.NoCommanders, pv.Settings)
	}
}

func TestFormatChangeLogLine(t *testing.T) {
	text := renderSettingsText(LogEvent{Seat: NoSeat, Label: game.SettingFormat, Choice: "standard"})
	if text != "The admin set the format to Standard" {
		t.Errorf("log line = %q", text)
	}
}
