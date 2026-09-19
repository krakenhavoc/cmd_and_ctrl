package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestSettingsVisibleToEveryViewer — ADR 0075 §2.2: the table's
// settings are public, so a seat, the other seat and a spectator all
// read the same object, and the legacy undo_limit mirrors it.
func TestSettingsVisibleToEveryViewer(t *testing.T) {
	g := newTwoSeatGame(t)
	limit, on := game.UndoUnlimited, true
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{UndoLimit: &limit, AllowSpawn: &on}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	want := TableSettingsView{
		UndoLimit:       game.UndoUnlimited,
		UndoScope:       "own",
		StartingLife:    game.StartingLife,
		CommanderDamage: game.CommanderDamageLethal,
		BotPace:         "normal",
		AllowSpawn:      true,
	}
	for _, viewer := range []string{g.Seats[0].ID.String(), g.Seats[1].ID.String(), ""} {
		v := ViewOfGameFor(g, viewer)
		if v.Settings == nil || *v.Settings != want {
			t.Errorf("viewer %q settings = %+v, want %+v", viewer, v.Settings, want)
		}
		if v.UndoLimit != game.UndoUnlimited {
			t.Errorf("viewer %q undo_limit = %d, want %d", viewer, v.UndoLimit, game.UndoUnlimited)
		}
	}

	raw, err := json.Marshal(ViewOfGameFor(g, g.Seats[0].ID.String()))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, key := range []string{`"settings":{`, `"undo_limit":-1`, `"undo_scope":"own"`,
		`"starting_life":40`, `"commander_damage":21`, `"bot_pace":"normal"`, `"allow_spawn":true`} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("wire view missing %s", key)
		}
	}
}
