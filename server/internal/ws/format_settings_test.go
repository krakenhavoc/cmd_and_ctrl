package ws

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The WebSocket settings action never changes a table's format (ADR
// 0222 §1): it has no seat list to ask whether a deck is loaded. A patch
// that names the table's own format, as a client echoing the settings it
// read does, is not a change.
func TestSettingsActionNeverChangesTheFormat(t *testing.T) {
	commander := game.DefaultTableSettings()
	standard := commander
	standard.Format = game.FormatStandard

	action := func(params string) actions.Action {
		return actions.Action{Type: actions.TypeSetTableSettings, Params: json.RawMessage(params)}
	}
	cases := []struct {
		name    string
		current game.TableSettings
		params  string
		want    bool
	}{
		{"no format named", commander, `{"undo_limit":3}`, false},
		{"commander at a commander table", commander, `{"format":"commander"}`, false},
		{"empty format at a commander table", commander, `{"format":""}`, false},
		{"standard at a commander table", commander, `{"format":"standard"}`, true},
		{"standard at a standard table", standard, `{"format":"standard","undo_limit":2}`, false},
		{"commander at a standard table", standard, `{"format":"commander"}`, true},
		{"unparseable is left to Dispatch", commander, `{"format":7}`, false},
	}
	for _, tc := range cases {
		if got := changesFormat(tc.current, action(tc.params)); got != tc.want {
			t.Errorf("%s: changesFormat = %v, want %v", tc.name, got, tc.want)
		}
	}
	other := actions.Action{Type: actions.TypeSetUndoLimit, Params: json.RawMessage(`{"format":"standard"}`)}
	if changesFormat(commander, other) {
		t.Error("an action that is not set_table_settings was read as a format change")
	}
}
