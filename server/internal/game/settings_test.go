package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// settings_test.go covers ADR 0075 sub-PR 1: the engine's table
// settings, their readers, the snapshot migration and the undo carve-
// out.

func intPtr(n int) *int { return &n }

// newLobbyGame seats `seats` players and does not start.
func newLobbyGame(t *testing.T, seats int) *Game {
	t.Helper()
	g := NewGame()
	for i := 0; i < seats; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck(fmt.Sprintf("Commander %d", i+1))); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	return g
}

func startLobbyGame(t *testing.T, g *Game) {
	t.Helper()
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
}

func settingsEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventSettingsChanged {
			out = append(out, ev)
		}
	}
	return out
}

func TestNewGameHasDefaultSettings(t *testing.T) {
	g := NewGame()
	if g.Settings != DefaultTableSettings() {
		t.Fatalf("NewGame settings = %+v, want %+v", g.Settings, DefaultTableSettings())
	}
	d := DefaultTableSettings()
	if d.UndoLimit != DefaultUndoLimit || d.StartingLife != StartingLife ||
		d.CommanderDamage != CommanderDamageLethal || d.UndoScope != UndoScopeOwn ||
		d.BotPace != BotPaceNormal || d.AllowSpawn {
		t.Errorf("defaults drifted from the constants: %+v", d)
	}
}

// The trap ADR 0075 §1 names: Start used to rewrite UndoLimit <= 0 to
// the default, so a table that asked for no undos got one.
func TestDeliberateZeroUndoLimitSurvivesStart(t *testing.T) {
	g := newLobbyGame(t, 2)
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{UndoLimit: intPtr(0)}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	startLobbyGame(t, g)
	if g.Settings.UndoLimit != 0 {
		t.Fatalf("UndoLimit after Start = %d, want 0", g.Settings.UndoLimit)
	}
	for i, p := range g.Seats {
		if p.UndosRemaining != 0 {
			t.Errorf("seat %d UndosRemaining = %d, want 0", i, p.UndosRemaining)
		}
	}
	if err := g.SpendUndo(g.Seats[0].ID); !errors.Is(err, ErrNoUndosRemaining) {
		t.Errorf("SpendUndo with a 0 limit = %v, want ErrNoUndosRemaining", err)
	}
}

func TestDeliberateZeroUndoLimitSurvivesSnapshotRoundTrip(t *testing.T) {
	g := newRestorableGame(t)
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{UndoLimit: intPtr(0)}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	_, restored := roundTrip(t, g)
	if restored.Settings.UndoLimit != 0 {
		t.Fatalf("restored UndoLimit = %d, want 0", restored.Settings.UndoLimit)
	}
	if restored.Settings != g.Settings {
		t.Errorf("restored settings = %+v, want %+v", restored.Settings, g.Settings)
	}
}

func TestEverySettingSurvivesSnapshotRoundTrip(t *testing.T) {
	g := newRestorableGame(t)
	scope, pace, on := UndoScopeHostAny, BotPaceSlow, true
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{
		UndoLimit: intPtr(UndoUnlimited), UndoScope: &scope,
		CommanderDamage: intPtr(15), BotPace: &pace, AllowSpawn: &on,
	}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	snap, restored := roundTrip(t, g)
	if snap.Schema != SnapshotSchemaVersion || SnapshotSchemaVersion < settingsSchemaVersion {
		t.Fatalf("capture schema %d, settings need >= %d", snap.Schema, settingsSchemaVersion)
	}
	if restored.Settings != g.Settings {
		t.Errorf("restored settings = %+v, want %+v", restored.Settings, g.Settings)
	}
}

// legacySnapshot rewrites a fresh capture into the pre-v4 shape: no
// settings object, the undo limit in the old top-level field.
func legacySnapshot(t *testing.T, g *Game, undoLimit int) *GameSnapshot {
	t.Helper()
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	delete(m, "settings")
	m["schema"] = json.RawMessage("3")
	m["undoLimit"] = json.RawMessage(fmt.Sprint(undoLimit))
	raw, err = json.Marshal(m)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	var s GameSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("decode legacy: %v", err)
	}
	return &s
}

func TestOldSchemaSnapshotMigratesToDefaults(t *testing.T) {
	cases := []struct {
		name      string
		lobby     bool
		undoLimit int
		want      int
	}{
		// A lobby game's 0 is the unset field the v3 Start would have
		// rewritten to the default.
		{"lobby unset limit", true, 0, DefaultUndoLimit},
		// An active v3 game's 0 was set on purpose (Start had already
		// rewritten any unset value), so it survives.
		{"active deliberate zero", false, 0, 0},
		{"active raised limit", false, 3, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var g *Game
			if tc.lobby {
				g = newLobbyGame(t, 2)
			} else {
				g = newRestorableGame(t)
			}
			s := legacySnapshot(t, g, tc.undoLimit)
			restored, err := s.Restore()
			if err != nil {
				t.Fatalf("Restore: %v", err)
			}
			want := DefaultTableSettings()
			want.UndoLimit = tc.want
			if restored.Settings != want {
				t.Errorf("migrated settings = %+v, want %+v", restored.Settings, want)
			}
		})
	}
}

func TestUnlimitedUndoNeverDebits(t *testing.T) {
	g := newActiveGame(t)
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{UndoLimit: intPtr(UndoUnlimited)}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	p := g.Seats[0]
	before := p.UndosRemaining
	for i := 0; i < 5; i++ {
		if !g.HasUndoBudget(p.ID) {
			t.Fatalf("HasUndoBudget false on undo %d under an unlimited budget", i+1)
		}
		if err := g.SpendUndo(p.ID); err != nil {
			t.Fatalf("SpendUndo %d: %v", i+1, err)
		}
	}
	if p.UndosRemaining != before {
		t.Errorf("UndosRemaining moved from %d to %d under an unlimited budget", before, p.UndosRemaining)
	}
}

func TestLegacySetUndoLimitClampsNegativeToZero(t *testing.T) {
	g := newActiveGame(t)
	if err := g.SetUndoLimit(g.Seats[0].ID, UndoUnlimited); err != nil {
		t.Fatalf("SetUndoLimit: %v", err)
	}
	if g.Settings.UndoLimit != 0 {
		t.Errorf("legacy SetUndoLimit(-1) = %d, want 0 (it must not select unlimited)", g.Settings.UndoLimit)
	}
	if err := g.SetUndoLimit(g.Seats[0].ID, 3); err != nil {
		t.Fatalf("SetUndoLimit: %v", err)
	}
	for i, p := range g.Seats {
		if p.UndosRemaining != 3 {
			t.Errorf("seat %d UndosRemaining = %d, want 3", i, p.UndosRemaining)
		}
	}
	if err := newLobbyGame(t, 2).SetUndoLimit(uuid.Nil, 2); !errors.Is(err, ErrGameNotActive) {
		t.Errorf("legacy SetUndoLimit in the lobby = %v, want ErrGameNotActive", err)
	}
}

func TestStartingLifeAppliesBeforeStartAndIsRejectedAfter(t *testing.T) {
	g := newLobbyGame(t, 2)
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{StartingLife: intPtr(30)}); err != nil {
		t.Fatalf("UpdateSettings in lobby: %v", err)
	}
	// A seat that joins after the change gets the new value too.
	if _, err := g.AddPlayer("P3", buildTestDeck("Commander 3")); err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	startLobbyGame(t, g)
	for i, p := range g.Seats {
		if p.Life != 30 {
			t.Errorf("seat %d life at start = %d, want 30", i, p.Life)
		}
	}

	err := g.UpdateSettings(uuid.Nil, SettingsPatch{StartingLife: intPtr(20), UndoLimit: intPtr(4)})
	if !errors.Is(err, ErrStartingLifeLocked) {
		t.Fatalf("StartingLife change after Start = %v, want ErrStartingLifeLocked", err)
	}
	if g.Settings.StartingLife != 30 || g.Settings.UndoLimit != DefaultUndoLimit {
		t.Errorf("a refused patch applied something: %+v", g.Settings)
	}
	// Re-sending the current value is not a change and is accepted, so
	// a client can post back the whole settings object mid-game.
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{StartingLife: intPtr(30)}); err != nil {
		t.Errorf("unchanged StartingLife after Start = %v, want nil", err)
	}
}

func TestSettingsPatchValidation(t *testing.T) {
	bad := UndoScope("everyone")
	pace := BotPace("ludicrous")
	for name, p := range map[string]SettingsPatch{
		"undo below unlimited": {UndoLimit: intPtr(-2)},
		"life 0":               {StartingLife: intPtr(0)},
		"life 1000":            {StartingLife: intPtr(1000)},
		"cmdr dmg 0":           {CommanderDamage: intPtr(0)},
		"cmdr dmg 100":         {CommanderDamage: intPtr(100)},
		"scope":                {UndoScope: &bad},
		"pace":                 {BotPace: &pace},
	} {
		g := newLobbyGame(t, 2)
		before := g.Settings
		if err := g.UpdateSettings(uuid.Nil, p); !errors.Is(err, ErrInvalidSetting) {
			t.Errorf("%s: err = %v, want ErrInvalidSetting", name, err)
		}
		if g.Settings != before || len(settingsEvents(g)) != 0 {
			t.Errorf("%s: an invalid patch changed state", name)
		}
	}
	g := newLobbyGame(t, 2)
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{
		UndoLimit: intPtr(UndoUnlimited), StartingLife: intPtr(999), CommanderDamage: intPtr(1),
	}); err != nil {
		t.Errorf("boundary values rejected: %v", err)
	}
}

func TestLoweredCommanderDamageKillsAtNextSBACheck(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	victim := g.Seats[1]
	cmdr := uuid.New()
	victim.RecordCommanderDamage(cmdr, 12)

	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if victim.Eliminated {
		t.Fatal("12 commander damage eliminated a player under the default threshold")
	}

	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{CommanderDamage: intPtr(10)}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if victim.Eliminated {
		t.Fatal("the settings change itself eliminated the player; it should wait for the SBA check")
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if !victim.Eliminated {
		t.Error("12 commander damage did not eliminate the player at a threshold of 10")
	}
}

func TestSettingsChangeSurvivesUndoRestore(t *testing.T) {
	g := newActiveGame(t)
	pre := g.Clone()
	on := true
	if err := g.UpdateSettings(g.Seats[0].ID, SettingsPatch{UndoLimit: intPtr(0), AllowSpawn: &on}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	want := g.Settings
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if g.Settings != want {
		t.Errorf("settings after undo = %+v, want the live %+v", g.Settings, want)
	}
}

func TestUpdateSettingsEmitsEventPerChangedField(t *testing.T) {
	g := newLobbyGame(t, 2)
	actor := g.Seats[0].ID
	on := true
	if err := g.UpdateSettings(actor, SettingsPatch{
		UndoLimit: intPtr(3), AllowSpawn: &on, CommanderDamage: intPtr(CommanderDamageLethal),
	}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	evs := settingsEvents(g)
	if len(evs) != 2 {
		t.Fatalf("got %d settings events, want 2 (the unchanged commander damage emits none): %+v", len(evs), evs)
	}
	want := map[string][2]string{
		SettingUndoLimit:  {"1", "3"},
		SettingAllowSpawn: {"false", "true"},
	}
	for _, ev := range evs {
		w, ok := want[ev.Label]
		if !ok {
			t.Errorf("unexpected setting %q", ev.Label)
			continue
		}
		if ev.Actor != actor || ev.SettingOld != w[0] || ev.SettingNew != w[1] {
			t.Errorf("%s event = actor %v %q→%q, want actor %v %q→%q",
				ev.Label, ev.Actor, ev.SettingOld, ev.SettingNew, actor, w[0], w[1])
		}
	}
}

func TestUndoLimitChangeRefreshesEverySeat(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	g.Seats[0].UndosRemaining = 0
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{UndoLimit: intPtr(2)}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	for i, p := range g.Seats {
		if p.UndosRemaining != 2 {
			t.Errorf("seat %d UndosRemaining = %d, want 2", i, p.UndosRemaining)
		}
	}
}

func TestCloneCopiesSettings(t *testing.T) {
	g := newActiveGame(t)
	scope := UndoScopeHostAny
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{UndoLimit: intPtr(5), UndoScope: &scope}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	c := g.Clone()
	if c.Settings != g.Settings {
		t.Fatalf("clone settings = %+v, want %+v", c.Settings, g.Settings)
	}
	// A value copy: changing the clone does not reach the original.
	c.Settings.UndoLimit = 9
	if g.Settings.UndoLimit != 5 {
		t.Error("the clone's settings alias the original's")
	}
}
