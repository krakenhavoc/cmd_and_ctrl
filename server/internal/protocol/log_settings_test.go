package protocol

// log_settings_test.go — ADR 0075 §2.3. Every settings change is
// narrated, because the settings are the rules the table agreed to
// play under. Until this file, EventSettingsChanged was the one
// TEMPORARY row in silentEventKinds.

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// settingsLines returns the settings entries of the public log, in
// order, as one viewer sees them.
func settingsLines(t *testing.T, g *game.Game, viewer string) []LogEvent {
	t.Helper()
	var out []LogEvent
	for _, e := range ViewOfGameFor(g, viewer).Log {
		if e.Kind == LogSettings {
			out = append(out, e)
		}
	}
	return out
}

func TestSettingsChangeIsNarratedForTheWholeTable(t *testing.T) {
	g := newTwoSeatGame(t)
	host := g.Seats[0]
	limit, spawn := 3, true
	if err := g.UpdateSettings(host.ID, game.SettingsPatch{UndoLimit: &limit, AllowSpawn: &spawn}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	// One line per field that moved, in the order UpdateSettings
	// emits them, and the same lines for every seat — settings are
	// public.
	for _, viewer := range []string{host.ID.String(), g.Seats[1].ID.String(), ""} {
		lines := settingsLines(t, g, viewer)
		if len(lines) != 2 {
			t.Fatalf("viewer %q: %d settings lines, want 2: %v", viewer, len(lines), texts(lines))
		}
		if got, want := lines[0].Text, host.Name+" (host) set undos to 3 per turn"; got != want {
			t.Errorf("viewer %q undo line = %q, want %q", viewer, got, want)
		}
		if got, want := lines[1].Text, host.Name+" (host) allowed spawning"; got != want {
			t.Errorf("viewer %q spawn line = %q, want %q", viewer, got, want)
		}
		// The structured half is what a client filters and renders a
		// control from (ADR 0075 sub-PR 5).
		if lines[0].Label != game.SettingUndoLimit || lines[0].Choice != "3" {
			t.Errorf("viewer %q undo line fields = %q/%q", viewer, lines[0].Label, lines[0].Choice)
		}
		if lines[1].Label != game.SettingAllowSpawn || lines[1].Choice != "true" {
			t.Errorf("viewer %q spawn line fields = %q/%q", viewer, lines[1].Label, lines[1].Choice)
		}
		for _, e := range lines {
			assertNoUUID(t, e.Text)
		}
	}
}

// TestSettingsChangeByTheAdminNamesNoSeat — an admin change has no
// seat behind it (Actor is uuid.Nil), and the line says so rather
// than blaming "someone" or the seat at index 0.
func TestSettingsChangeByTheAdminNamesNoSeat(t *testing.T) {
	g := newTwoSeatGame(t)
	off := false
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{AllowSpawn: &off}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	// AllowSpawn is already false, so nothing changed and nothing is
	// said — a no-op patch emits no event.
	if lines := settingsLines(t, g, ""); len(lines) != 0 {
		t.Fatalf("a no-op patch was narrated: %v", texts(lines))
	}

	on := true
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{AllowSpawn: &on}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	lines := settingsLines(t, g, "")
	if len(lines) != 1 {
		t.Fatalf("%d settings lines, want 1: %v", len(lines), texts(lines))
	}
	if got, want := lines[0].Text, "The admin allowed spawning"; got != want {
		t.Errorf("line = %q, want %q", got, want)
	}
	if lines[0].Seat != NoSeat {
		t.Errorf("seat = %d, want NoSeat", lines[0].Seat)
	}
}

// TestSettingsSentences walks every key the engine can emit, so a new
// setting whose sentence nobody wrote shows up here as "set X to Y"
// rather than as a silence.
func TestSettingsSentences(t *testing.T) {
	g := newTwoSeatGame(t)
	host := g.Seats[0]
	unlimited, none, one, four := game.UndoUnlimited, 0, 1, 4
	scope := game.UndoScopeHostAny
	own := game.UndoScopeOwn
	dmg := 15
	pace := game.BotPaceSlow
	off := false

	steps := []struct {
		patch game.SettingsPatch
		want  string
	}{
		{game.SettingsPatch{UndoLimit: &unlimited}, "made undos unlimited"},
		{game.SettingsPatch{UndoLimit: &none}, "turned undos off"},
		{game.SettingsPatch{UndoLimit: &one}, "set undos to 1 per turn"},
		{game.SettingsPatch{UndoLimit: &four}, "set undos to 4 per turn"},
		{game.SettingsPatch{UndoScope: &scope}, "may now undo anyone's action"},
		{game.SettingsPatch{UndoScope: &own}, "limited undo to each player's own actions"},
		{game.SettingsPatch{CommanderDamage: &dmg}, "set lethal commander damage to 15"},
		{game.SettingsPatch{BotPace: &pace}, "set the bots' pace to slow"},
	}
	for i, s := range steps {
		if err := g.UpdateSettings(host.ID, s.patch); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
	// And the off half of the switch, which the on half above did not
	// reach (AllowSpawn starts false).
	on := true
	if err := g.UpdateSettings(host.ID, game.SettingsPatch{AllowSpawn: &on}); err != nil {
		t.Fatal(err)
	}
	if err := g.UpdateSettings(host.ID, game.SettingsPatch{AllowSpawn: &off}); err != nil {
		t.Fatal(err)
	}

	lines := settingsLines(t, g, "")
	want := make([]string, 0, len(steps)+2)
	for _, s := range steps {
		want = append(want, s.want)
	}
	want = append(want, "allowed spawning", "disallowed spawning")
	if len(lines) != len(want) {
		t.Fatalf("%d lines, want %d: %v", len(lines), len(want), texts(lines))
	}
	for i, w := range want {
		if !strings.HasPrefix(lines[i].Text, host.Name+" (host) ") || !strings.HasSuffix(lines[i].Text, w) {
			t.Errorf("line %d = %q, want %q … %q", i, lines[i].Text, host.Name+" (host)", w)
		}
	}
}

// TestStartingLifeSentence needs a table that has not started — the
// setting is locked once it has (ADR 0075 §2.3), so its line can only
// be written in the lobby.
func TestStartingLifeSentence(t *testing.T) {
	g := game.NewGame()
	deck := make([]game.Card, 10)
	for i := range deck {
		deck[i] = game.NewCard("filler", uuid.Nil)
	}
	if _, err := g.AddPlayer("Alice", deck); err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	host := g.Seats[0]
	life := 30
	if err := g.UpdateSettings(host.ID, game.SettingsPatch{StartingLife: &life}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	lines := settingsLines(t, g, "")
	if len(lines) != 1 {
		t.Fatalf("%d lines, want 1: %v", len(lines), texts(lines))
	}
	if got, want := lines[0].Text, host.Name+" (host) set starting life to 30"; got != want {
		t.Errorf("line = %q, want %q", got, want)
	}
}

func texts(entries []LogEvent) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Text)
	}
	return out
}
